/*******************************************************************************
** @Author:					Thomas Bouder <Tbouder>
** @Email:					Tbouder@protonmail.com
** @Date:					Tuesday 07 January 2020 - 14:13:47
** @Filename:				service.go
**
** @Last modified by:		Tbouder
** @Last modified time:		Sunday 09 February 2020 - 17:41:25
*******************************************************************************/

package			main

import (
	"io"
	"io/ioutil"
	"context"
	"strings"
	"time"
	"strconv"
	"github.com/microgolang/logs"
	"github.com/google/uuid"
	P "github.com/microgolang/postgre"
)

func	CreatePictureRef(req *UploadPictureRequest, blob []byte, GroupID, size string, oW, oH uint) (int, int, string, error) {
	var	width int
	var	height int
	var	path string
	var	encryptionKey string
	var	err error

	if (size == `original`) {
		width, height, path, encryptionKey, err = createOriginal(blob, req.GetContent().GetType(), req.GetMemberID())
	} else {
		width, height, path, encryptionKey, err = createThumbnails(blob, req.GetContent().GetType(), req.GetMemberID(), size, oW, oH)
	}

	if (err != nil) {
		return -1, -1, ``, err
	}

	unixTimeStamp, _ := strconv.ParseInt(req.GetContent().GetOriginalTime(), 10, 64)
	timeFormated := time.Unix(0, unixTimeStamp * int64(time.Millisecond)).Format(`2006-01-02 15:04:05`)
	toInsert := []P.S_InsertorWhere{
		P.S_InsertorWhere{Key: `GroupID`, Value: GroupID},
		P.S_InsertorWhere{Key: `MemberID`, Value: req.GetMemberID()},
		P.S_InsertorWhere{Key: `Name`, Value: req.GetContent().GetName()},
		P.S_InsertorWhere{Key: `Type`, Value: req.GetContent().GetType()},
		P.S_InsertorWhere{Key: `EncryptionKey`, Value: encryptionKey},
		P.S_InsertorWhere{Key: `Path`, Value: path},
		P.S_InsertorWhere{Key: `OriginalTime`, Value: timeFormated},
		P.S_InsertorWhere{Key: `Size`, Value: size},
		P.S_InsertorWhere{Key: `Width`, Value: strconv.Itoa(width)},
		P.S_InsertorWhere{Key: `Height`, Value: strconv.Itoa(height)},
		P.S_InsertorWhere{Key: `Weight`, Value: strconv.FormatInt(int64(len(blob)), 10)},
	}
	if (req.GetAlbumID() != ``) {
		toInsert = append(toInsert, P.S_InsertorWhere{Key: `AlbumID`, Value: req.GetAlbumID()})
	}

	_, err = P.NewInsertor(PGR).Into(`pictures`).Values(toInsert...).Do()
	if (err != nil) {
		return -1, -1, ``, err
	}
	return width, height, timeFormated, nil
}
/******************************************************************************
**	UploadPicture
**	Request -> Stream
**	Response -> Standard
**
**	Receive a stream request from the proxy, containing the non-crypted image
**	data, recreate it, and send-it to the Keys microservice to encrypt it.
**	Once it's done, save the crypted file to the storage and add it to the
**	database.
**	TODO: Optimization by directly streaming the received stream to the
**	ms, without rebuilding the image first.
**
**	UploadPicture => version which stream the stream to the key MS
******************************************************************************/
func (s *server) UploadPicture(stream PicturesService_UploadPictureServer) error {
	req := &UploadPictureRequest{}

	for {
		select {
			case <-stream.Context().Done():
				return nil
			default:
		}

		recv, err := stream.Recv()
		if err == io.EOF {
			GroupID := uuid.New().String()
			/******************************************************************
			**	Create the reference for the 500x500 picture in the Database
			******************************************************************/
			width, height, originalTime, err := CreatePictureRef(req, req.GetChunk(), GroupID, `500x500`, 500, 0)
			if (err != nil) {
				stream.Send(&UploadPictureResponse{Step: 4, Success: false})
				return err
			}

			/******************************************************************
			**	Create the different thumbnails for this image :
			**	1000 x 1000 => for access
			**	original => for better
			******************************************************************/
			go CreatePictureRef(req, req.GetChunk(), GroupID, `1000x1000`, 1000, 0)
			go CreatePictureRef(req, req.GetChunk(), GroupID, `original`, 0, 0)

			/******************************************************************
			**	Send a message to the websocket to inform the client the image
			**	is now in the database.
			******************************************************************/
			stream.Send(&UploadPictureResponse{
				Step: 4,
				Picture: &ListPictures_Content{
					Uri: GroupID,
					OriginalTime: originalTime,
					Width: uint32(width),
					Height: uint32(height),
				},
				Success: true,
			})
			stream.Context().Done()
			return (nil)
		}
		if err != nil {
			logs.Error("receive error : ", err)
			continue
		}

		/**********************************************************************
		**	Use this for direct streaming
		**********************************************************************/
		req.MemberID = recv.GetMemberID()
		req.Content = recv.GetContent()
		req.AlbumID = recv.GetAlbumID()
		req.Chunk = append(req.GetChunk(), recv.GetChunk()...)
	}
}

/******************************************************************************
**	DownloadPicture
**	Request -> Standard
**	Response -> Stream
**
**	Take a request, containing the information to identify the image and the
**	member asking for access, access the encrypted image, decrypt it, and
**	stream it as a response to the proxy.
******************************************************************************/
func (s *server) DownloadPicture(req *DownloadPictureRequest, stream PicturesService_DownloadPictureServer) (error) {
	var	Width int
	var	Height int
	var	Path string
	var	Type string
	var	EncryptionKey string
	var	MemberID string
	/**************************************************************************
	**	0. Get the information about the picture we try to access
	**************************************************************************/
	err := P.NewSelector(PGR).Select(`Path`, `Type`, `EncryptionKey`, `Width`, `Height`, `MemberID`).From(`pictures`).Where(
		P.S_SelectorWhere{Key: `GroupID`, Value: req.GetPictureID()},
		P.S_SelectorWhere{Key: `Size`, Value: req.GetPictureSize()},
	).One(&Path, &Type, &EncryptionKey, &Width, &Height, &MemberID)
	if (err != nil) {
		logs.Error(`Impossible to get image`, err)
		return err
	}

	/**************************************************************************
	**	1. Read the file and store it's encrypted content, as []byte, to a
	**	variable which we will send to the key microservice to decrypt it
	**************************************************************************/
	encryptedData, err := ioutil.ReadFile(Path)
	if (err != nil) {
		logs.Error(`Impossible to read image`, err)
		return err
	}

	/**************************************************************************
	**	2. Send the MemberID, the encrypted data, and the picture decryption
	**	key to the pictures microservice to get the decrypted data.
	**************************************************************************/
	isSuccess, decryptedData, err := DecryptPicture(MemberID, encryptedData, EncryptionKey, req.GetHashKey())
	if (err != nil || !isSuccess) {
		logs.Error(`Impossible to decrypt image ` + err.Error())
		return nil
	}

	/**************************************************************************
	**	3. Chunk the file according to DEFAULT_CHUNK_SIZE (64 * 1000) and send
	**	back the full message to the proxy
	**************************************************************************/
	fileSize := len(decryptedData)
	resp := &DownloadPictureResponse{ContentType: Type, Width: uint32(Width), Height: uint32(Height)}

	for currentByte := 0; currentByte < fileSize; currentByte += DEFAULT_CHUNK_SIZE {
		if currentByte + DEFAULT_CHUNK_SIZE > fileSize {
			resp.Chunk = decryptedData[currentByte:fileSize]
		} else {
			resp.Chunk = decryptedData[currentByte : currentByte + DEFAULT_CHUNK_SIZE]
		}
		if err := stream.Send(resp); err != nil {
			logs.Error(err)
			return err
		}
	}

	return nil
}

/******************************************************************************
**	DeletePicture
**	Request -> Standard
**	Response -> Stream
**
**	Take a request, containing the information to identify the image and the
**	member asking for access, access the encrypted image, decrypt it, and
**	stream it as a response to the proxy.
******************************************************************************/
func (s *server) DeletePictures(ctx context.Context, req *DeletePicturesRequest) (*DeletePicturesResponse, error) {
	for _, pictureID := range req.GetPicturesID() {
		type myReturnType struct {Path string}
		pictureID = strings.Split(pictureID, `?`)[0]

		rows, err := P.NewSelector(PGR).
			From(`pictures`).
			Select(`Path`).
			Where(P.S_SelectorWhere{Key: `GroupID`, Value: pictureID}).
			All(&[]myReturnType{})

		if (err != nil) {
			logs.Error(`Impossible to find the path for this groupd`, err)
			return &DeletePicturesResponse{Success: false}, err
		}
	
		assertedRows := rows.([]myReturnType)
		for _, row := range assertedRows {removePicture(row.Path)}

		P.NewDeletor(PGR).Into(`pictures`).Where(P.S_DeletorWhere{Key: `GroupID`, Value: pictureID}).Do()
	}

	return &DeletePicturesResponse{Success: true}, nil
}

/******************************************************************************
**	ListPicturesByMemberID
******************************************************************************/
func (s *server) ListPicturesByMemberID(ctx context.Context, req *ListPicturesByMemberIDRequest) (*ListPicturesByMemberIDResponse, error) {
	type myReturnType struct {
		Width int
		Height int
		GroupID string
		OriginalTime string
		Day string
	}
	var response [](*ListPictures_Content)

	rows, err := P.NewSelector(PGR).
	Select(`Width`, `Height`, `GroupID`, `OriginalTime`, `date_trunc('day', OriginalTime) as Day`).
	From(`pictures`).
	Where(
		P.S_SelectorWhere{Key: `MemberID`, Value: req.GetMemberID()},
		P.S_SelectorWhere{Key: `Size`, Value: `500x500`},
	).
	Sort(`OriginalTime`, `DESC`).
	All(&[]myReturnType{})
	if (err != nil) {
		logs.Error(`Impossible to get images`, err)
		return &ListPicturesByMemberIDResponse{Pictures: response}, err
	}


	var alt = make(map[string](*ListPictures_Wrapper))


	assertedRows := rows.([]myReturnType)
	for _, row := range assertedRows {
		response = append(response, &ListPictures_Content{
			Uri: row.GroupID,
			OriginalTime: row.OriginalTime,
			Width: uint32(row.Width),
			Height: uint32(row.Height),
		})

		if _, ok := alt[row.Day]; ok == true {
			alt[row.Day].PicturesAlt = append(alt[row.Day].PicturesAlt, &ListPictures_Content{
				Uri: row.GroupID,
				OriginalTime: row.OriginalTime,
				Width: uint32(row.Width),
				Height: uint32(row.Height),
			})
		} else {
			newAlt := new(ListPictures_Wrapper)
			newAlt.PicturesAlt = make([](*ListPictures_Content), 0)
			newAlt.PicturesAlt = append(newAlt.PicturesAlt, &ListPictures_Content{
				Uri: row.GroupID,
				OriginalTime: row.OriginalTime,
				Width: uint32(row.Width),
				Height: uint32(row.Height),
			})
			alt[row.Day] = newAlt
		}
	}


	return &ListPicturesByMemberIDResponse{Pictures: response, PicturesAlt: alt}, nil
}

/******************************************************************************
**	SetPictureAlbum
******************************************************************************/
func (s *server) SetPicturesAlbum(ctx context.Context, req *SetPicturesAlbumRequest) (*SetPicturesAlbumResponse, error) {
	err := P.NewUpdator(PGR).Set(
		P.S_UpdatorSetter{Key: `AlbumID`, Value: req.GetAlbumID()},
	).Where(
		P.S_UpdatorWhere{Key: `GroupID`, Action: `IN`, Values: req.GetGroupIDs()},
		P.S_UpdatorWhere{Key: `MemberID`, Value: req.GetMemberID()},
	).Into(`pictures`).Do()

	return &SetPicturesAlbumResponse{Success: err == nil}, err
}


/******************************************************************************
**	ListPicturesByAlbumID
******************************************************************************/
func (s *server) ListPicturesByAlbumID(ctx context.Context, req *ListPicturesByAlbumIDRequest) (*ListPicturesByAlbumIDResponse, error) {
	type myReturnType struct {
		Width int
		Height int
		GroupID string
		OriginalTime string
		Day string
	}
	var response [](*ListPictures_Content)

	rows, err := P.NewSelector(PGR).
	Select(`Width`, `Height`, `GroupID`, `OriginalTime`, `date_trunc('day', OriginalTime) as Day`).
	From(`pictures`).
	Where(
		P.S_SelectorWhere{Key: `AlbumID`, Value: req.GetAlbumID()},
		P.S_SelectorWhere{Key: `MemberID`, Value: req.GetMemberID()},
		P.S_SelectorWhere{Key: `Size`, Value: `500x500`},
	).
	Sort(`OriginalTime`, `DESC`).
	All(&[]myReturnType{})
	if (err != nil) {
		logs.Error(`Impossible to get images`, err)
		return &ListPicturesByAlbumIDResponse{Pictures: response}, err
	}

	assertedRows := rows.([]myReturnType)
	for _, row := range assertedRows {
		response = append(response, &ListPictures_Content{
			Uri: row.GroupID,
			OriginalTime: row.OriginalTime,
			Width: uint32(row.Width),
			Height: uint32(row.Height),
		})
	}

	return &ListPicturesByAlbumIDResponse{Pictures: response}, nil
}
