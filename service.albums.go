/*******************************************************************************
** @Author:					Thomas Bouder <Tbouder>
** @Email:					Tbouder@protonmail.com
** @Date:					Tuesday 14 January 2020 - 20:27:25
** @Filename:				service.albums.go
**
** @Last modified by:		Tbouder
** @Last modified time:		Saturday 15 February 2020 - 14:29:12
*******************************************************************************/


package			main

import (
	"strings"
	"context"
	"strconv"
	"github.com/microgolang/logs"
	"database/sql"
	"github.com/panghostlin/SDK/Pictures"
	P "github.com/microgolang/postgre"
)

/******************************************************************************
**	CreateAlbum
******************************************************************************/
func (s *server) CreateAlbum(ctx context.Context, req *pictures.CreateAlbumRequest) (*pictures.CreateAlbumResponse, error) {
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
	}

	return &pictures.CreateAlbumResponse{AlbumID: ID, Name: req.GetName()}, err
}

/******************************************************************************
**	GetAlbum
******************************************************************************/
func (s *server) GetAlbum(ctx context.Context, req *pictures.GetAlbumRequest) (*pictures.GetAlbumResponse, error) {
	var	response pictures.GetAlbumsResponse_Content

	err := P.NewSelector(PGR).
		From(`albums`).
		Select(`ID`, `Name`, `NumberOfPictures`).
		Where(
			P.S_SelectorWhere{Key: `MemberID`, Value: req.GetMemberID()},
			P.S_SelectorWhere{Key: `ID`, Value: req.GetAlbumID()},
		).
		One(&response.AlbumID, &response.Name, &response.NumberOfPictures)

	return &pictures.GetAlbumResponse{Album: &response}, err
}

/******************************************************************************
**	SetAlbumCover
******************************************************************************/
func (s *server) SetAlbumCover(ctx context.Context, req *pictures.SetAlbumCoverRequest) (*pictures.SetAlbumCoverResponse, error) {
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

	return &pictures.SetAlbumCoverResponse{AlbumID: req.GetAlbumID()}, err
}

/******************************************************************************
**	ListAlbums
******************************************************************************/
func (s *server) ListAlbums(ctx context.Context, req *pictures.ListAlbumsRequest) (*pictures.ListAlbumsResponse, error) {
	type myReturnType struct {
		ID string
		Name string
		NumberOfPictures int
		CoverPicture0ID sql.NullString
		CoverPicture1ID sql.NullString
		CoverPicture2ID sql.NullString
	}
	var	response []*pictures.ListAlbumsResponse_Content

	rows, err := P.NewSelector(PGR).
		From(`albums`).
		Select(`ID`, `Name`, `NumberOfPictures`, `CoverPicture0ID`, `CoverPicture1ID`, `CoverPicture2ID`).
		Where(P.S_SelectorWhere{Key: `MemberID`, Value: req.GetMemberID()}).
		Sort(`CreationTime`, `DESC`).
		All(&[]myReturnType{})
	if (err != nil) {
		logs.Error(`Impossible to get albums`, err)
		return &pictures.ListAlbumsResponse{Albums: response}, err
	}

	assertedRows := rows.([]myReturnType)
	for _, row := range assertedRows {
		response = append(response, &pictures.ListAlbumsResponse_Content{
			AlbumID: row.ID,
			Name: row.Name,
			NumberOfPictures: int32(row.NumberOfPictures),
			CoverPicture0ID: row.CoverPicture0ID.String,
			CoverPicture1ID: row.CoverPicture1ID.String,
			CoverPicture2ID: row.CoverPicture2ID.String,
		})
	}

	return &pictures.ListAlbumsResponse{Albums: response}, nil
}

/******************************************************************************
**	DeleteAlbum
******************************************************************************/
func (s *server) DeleteAlbum(ctx context.Context, req *pictures.DeleteAlbumRequest) (*pictures.DeleteAlbumResponse, error) {
	err := P.NewDeletor(PGR).
		Into(`albums`).
		Where(
			P.S_DeletorWhere{Key: `ID`, Value: req.GetAlbumID()},
			P.S_DeletorWhere{Key: `MemberID`, Value: req.GetMemberID()},
		).
		Do()

	return &pictures.DeleteAlbumResponse{Success: err == nil}, err
}

/******************************************************************************
**	SetAlbumName
******************************************************************************/
func (s *server) SetAlbumName(ctx context.Context, req *pictures.SetAlbumNameRequest) (*pictures.SetAlbumNameResponse, error) {
	err := P.NewUpdator(PGR).Set(
		P.S_UpdatorSetter{Key: `Name`, Value: req.GetName()},
	).Where(
		P.S_UpdatorWhere{Key: `ID`, Value: req.GetAlbumID()},
		P.S_UpdatorWhere{Key: `MemberID`, Value: req.GetMemberID()},
	).Into(`albums`).Do()

	return &pictures.SetAlbumNameResponse{AlbumID: req.GetAlbumID()}, err
}