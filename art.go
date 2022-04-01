package main

import (
	"database/sql"
	"fmt"
)

type ArtShow struct {
	ConnectedUsers int        `json:"connectedUsers"`
	IsExhibiting   bool       `json:"isExhibiting"`
	ID             int        `json:"id"`
	Name           string     `json:"name"`
	Description    string     `json:"description"`
	Image          string     `json:"image"`
	Artists        Artist     `json:"artist"` //should be n to n but for semplicity we are gonna say it's a 1 to 1
	Pieces         []ArtPiece `json:"pieces"`
	Connected      []*Client  `json:"-"`
}

//it will retrive all the pieces of the art show from the database
func (a *ArtShow) GetPieces(connection *sql.DB, wantLikes bool) error {
	getPiecesQuery := `
		SELECT ap.*
		FROM artpieces AS ap
		JOIN artshows AS ash ON ap.idArtShow = ash.id
		WHERE ash.id = ?`
	results, err := connection.Query(getPiecesQuery, a.ID)
	if err != nil {
		return err
	}
	defer results.Close()

	a.Pieces = []ArtPiece{}
	for results.Next() {
		var piece ArtPiece
		err := results.Scan(&piece.ID, &piece.IDArtShow, &piece.IDArtist, &piece.Title, &piece.Year, &piece.Dimensions, &piece.Technics, &piece.Image)
		if err != nil {
			return err
		}
		if wantLikes {
			piece.GetLikes(connection, a.ID)
		}
		a.Pieces = append(a.Pieces, piece)
	}
	return nil
}

func (a *ArtShow) AddClient(client *Client) {
	a.Connected = append(a.Connected, client)
}

func (a *ArtShow) AddLike(connection *sql.DB, userID, pieceID int) error {
	insertLikeQuery := `INSERT INTO likes (idArtShow, idArtPiece, idUser) VALUES (?, ?, ?)`
	_, err := connection.Exec(insertLikeQuery, a.ID, pieceID, userID)
	if err != nil {
		return err
	}
	//update the likes count of the piece
	for i := range a.Pieces {
		if a.Pieces[i].ID == pieceID {
			a.Pieces[i].Likes++
		}
	}
	return nil
}

func (a *ArtShow) RemoveLike(connection *sql.DB, userID, pieceID int) error {
	deleteLikeQuery := `DELETE FROM likes WHERE idArtShow = ? AND idArtPiece = ? AND idUser = ?`
	_, err := connection.Exec(deleteLikeQuery, a.ID, pieceID, userID)
	if err != nil {
		return err
	}
	//update the likes count of the piece
	for i := range a.Pieces {
		if a.Pieces[i].ID == pieceID {
			a.Pieces[i].Likes--
		}
	}
	return nil
}

func (a *ArtShow) GetArtPiece(pieceID int, wantLike bool, connection *sql.DB) ArtPiece {
	for i := range a.Pieces {
		if a.Pieces[i].ID == pieceID {
			if wantLike {
				a.Pieces[i].GetLikes(connection, a.ID)
			}
			return a.Pieces[i]
		}
	}
	fmt.Println("piece not found")
	return ArtPiece{}
}

type ArtPiece struct {
	ID         int    `json:"id"`
	IDArtShow  int    `json:"idArtShow"`
	IDArtist   int    `json:"idArtist"`
	Title      string `json:"title"`
	Year       string `json:"year"`
	Dimensions string `json:"dimensions"`
	Technics   string `json:"technics"`
	Image      string `json:"image"`
	Likes      int    `json:"likes"`
}

func (a *ArtPiece) GetLikes(connection *sql.DB, artShowID int) error {
	getLikesQuery := `SELECT COUNT(*) FROM likes WHERE idArtPiece = ? AND idArtShow = ?`
	return connection.QueryRow(getLikesQuery, a.ID, artShowID).Scan(&a.Likes)
}

type Artist struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Image       string `json:"image"`
}

//get all art shows from the database
func GetArtShows(connection *sql.DB, getPieces bool) ([]ArtShow, error) {
	var artShows []ArtShow
	getAllArtShowQuery := `
		SELECT ash.id, ash.name, ash.description, ash.image, ar.id, ar.name, ar.description, ar.image FROM artshows AS ash INNER JOIN artists AS ar ON ash.artistID = ar.id`
	results, err := connection.Query(getAllArtShowQuery)
	if err != nil {
		return nil, err
	}
	defer results.Close()
	//iterator
	for results.Next() {
		var artShow ArtShow
		err := results.Scan(&artShow.ID, &artShow.Name, &artShow.Description, &artShow.Image, &artShow.Artists.ID, &artShow.Artists.Name, &artShow.Artists.Description, &artShow.Artists.Image)
		if err != nil {
			return nil, err
		}
		fmt.Println("artShow: ", artShow.Name)
		//get the pieces of the art show
		if getPieces {
			err = artShow.GetPieces(connection, false)
			if err != nil {
				return nil, err
			}
		}
		artShows = append(artShows, artShow)
	}
	return artShows, nil
}

//create a new user in the database and return the id of the user
func generateNewUser(connection *sql.DB) (int, error) {
	//insert into users
	lastInsertId := 0
	err := connection.QueryRow("INSERT INTO users (username) VALUES($1) RETURNING iduser", "ciao").Scan(&lastInsertId)
	//get the id of the user
	return lastInsertId, err
}

func userHasLiked(connection *sql.DB, userID, artShowID, artPieceID int) (bool, error) {
	var hasLiked bool
	getLikeQuery := `SELECT COUNT(*) FROM likes WHERE idUser = ? AND idArtShow = ? AND idArtPiece = ?`
	err := connection.QueryRow(getLikeQuery, userID, artShowID, artPieceID).Scan(&hasLiked)
	if err != nil {
		return false, err
	}
	return hasLiked, nil
}
