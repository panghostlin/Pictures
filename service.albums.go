/*******************************************************************************
** @Author:					Thomas Bouder <Tbouder>
** @Email:					Tbouder@protonmail.com
** @Date:					Tuesday 14 January 2020 - 20:27:25
** @Filename:				service.albums.go
**
** @Last modified by:		Tbouder
** @Last modified time:		Thursday 06 February 2020 - 15:40:24
*******************************************************************************/


package			main

import (
	"strings"
	"context"
	"strconv"
	"github.com/microgolang/logs"
	"database/sql"
	P "github.com/microgolang/postgre"
)

/******************************************************************************
**	CreateAlbum
******************************************************************************/
func (s *server) CreateAlbum(ctx context.Context, req *CreateAlbumRequest) (*CreateAlbumResponse, error) {
	coverPicture0ID := strings.Split(req.GetCoverPicture0ID(), `?`)[0]
	coverPicture1ID := strings.Split(req.GetCoverPicture1ID(), `?`)[0]
	coverPicture2ID := strings.Split(req.GetCoverPicture2ID(), `?`)[0]

	toInsert := []P.S_InsertorWhere{
		P.S_InsertorWhere{Key: `MemberID`, Value: req.GetMemberID()},
		P.S_InsertorWhere{Key: `Name`, Value: req.GetName()},
		P.S_InsertorWhere{Key: `NumberOfPictures`, Value: strconv.Itoa(0)},
	}
	if (coverPicture0ID != ``) {toInsert = append(toInsert, P.S_InsertorWhere{Key: `CoverPicture0ID`, Value: coverPicture0ID})}
	if (coverPicture1ID != ``) {toInsert = append(toInsert, P.S_InsertorWhere{Key: `CoverPicture1ID`, Value: coverPicture1ID})}
	if (coverPicture2ID != ``) {toInsert = append(toInsert, P.S_InsertorWhere{Key: `CoverPicture2ID`, Value: coverPicture2ID})}

	ID, err := P.NewInsertor(PGR).Into(`albums`).Values(toInsert...).Do()

	for _, eachPicture := range req.GetPictures() {
		eachPictureID := strings.Split(eachPicture, `?`)[0]
		err = P.NewUpdator(PGR).Set(P.S_UpdatorSetter{Key: `AlbumID`, Value: ID}).
			Where(P.S_UpdatorWhere{Key: `GroupID`, Value: eachPictureID}).
			Into(`pictures`).Do()
		logs.Pretty(err)
	}

	return &CreateAlbumResponse{AlbumID: ID, Name: req.GetName()}, err
}

/******************************************************************************
**	GetAlbum
******************************************************************************/
func (s *server) GetAlbum(ctx context.Context, req *GetAlbumRequest) (*GetAlbumResponse, error) {
	var	response GetAlbumsResponse_Content

	err := P.NewSelector(PGR).
		From(`albums`).
		Select(`ID`, `Name`, `NumberOfPictures`).
		Where(
			P.S_SelectorWhere{Key: `MemberID`, Value: req.GetMemberID()},
			P.S_SelectorWhere{Key: `ID`, Value: req.GetAlbumID()},
		).
		One(&response.AlbumID, &response.Name, &response.NumberOfPictures)

	return &GetAlbumResponse{Album: &response}, err
}

/******************************************************************************
**	SetAlbumCover
******************************************************************************/
func (s *server) SetAlbumCover(ctx context.Context, req *SetAlbumCoverRequest) (*SetAlbumCoverResponse, error) {
	coverPicture0ID := strings.Split(req.GetCoverPicture0ID(), `?`)[0]
	coverPicture1ID := strings.Split(req.GetCoverPicture1ID(), `?`)[0]
	coverPicture2ID := strings.Split(req.GetCoverPicture2ID(), `?`)[0]

	err := P.NewUpdator(PGR).Set(
		P.S_UpdatorSetter{Key: `CoverPicture0ID`, Value: coverPicture0ID},
		P.S_UpdatorSetter{Key: `CoverPicture1ID`, Value: coverPicture1ID},
		P.S_UpdatorSetter{Key: `CoverPicture2ID`, Value: coverPicture2ID},
	).Where(
		P.S_UpdatorWhere{Key: `ID`, Value: req.GetAlbumID()},
		P.S_UpdatorWhere{Key: `MemberID`, Value: req.GetMemberID()},
	).Into(`albums`).Do()

	return &SetAlbumCoverResponse{AlbumID: req.GetAlbumID()}, err
}

/******************************************************************************
**	ListAlbums
******************************************************************************/
func (s *server) ListAlbums(ctx context.Context, req *ListAlbumsRequest) (*ListAlbumsResponse, error) {
	type myReturnType struct {
		ID string
		Name string
		NumberOfPictures int
		CoverPicture0ID sql.NullString
		CoverPicture1ID sql.NullString
		CoverPicture2ID sql.NullString
	}
	var	response []*ListAlbumsResponse_Content

	rows, err := P.NewSelector(PGR).
		From(`albums`).
		Select(`ID`, `Name`, `NumberOfPictures`, `CoverPicture0ID`, `CoverPicture1ID`, `CoverPicture2ID`).
		Where(P.S_SelectorWhere{Key: `MemberID`, Value: req.GetMemberID()}).
		Sort(`CreationTime`, `DESC`).
		All(&[]myReturnType{})
	if (err != nil) {
		logs.Error(`Impossible to get albums`, err)
		return &ListAlbumsResponse{Albums: response}, err
	}

	assertedRows := rows.([]myReturnType)
	for _, row := range assertedRows {
		response = append(response, &ListAlbumsResponse_Content{
			AlbumID: row.ID,
			Name: row.Name,
			NumberOfPictures: int32(row.NumberOfPictures),
			CoverPicture0ID: row.CoverPicture0ID.String,
			CoverPicture1ID: row.CoverPicture1ID.String,
			CoverPicture2ID: row.CoverPicture2ID.String,
		})
	}

	return &ListAlbumsResponse{Albums: response}, nil
}

/******************************************************************************
**	DeleteAlbum
******************************************************************************/
func (s *server) DeleteAlbum(ctx context.Context, req *DeleteAlbumRequest) (*DeleteAlbumResponse, error) {
	err := P.NewDeletor(PGR).
		Into(`albums`).
		Where(
			P.S_DeletorWhere{Key: `ID`, Value: req.GetAlbumID()},
			P.S_DeletorWhere{Key: `MemberID`, Value: req.GetMemberID()},
		).
		Do()

	return &DeleteAlbumResponse{Success: err == nil}, err
}

/******************************************************************************
**	SetAlbumCover
******************************************************************************/
func (s *server) SetAlbumName(ctx context.Context, req *SetAlbumNameRequest) (*SetAlbumNameResponse, error) {
	err := P.NewUpdator(PGR).Set(
		P.S_UpdatorSetter{Key: `Name`, Value: req.GetName()},
	).Where(
		P.S_UpdatorWhere{Key: `ID`, Value: req.GetAlbumID()},
		P.S_UpdatorWhere{Key: `MemberID`, Value: req.GetMemberID()},
	).Into(`albums`).Do()

	return &SetAlbumNameResponse{AlbumID: req.GetAlbumID()}, err
}