package main

import (
	"fmt"
	"net/http"
)

type ArtShow struct {
	Type      string `json:"type"`
	Connected int    `json:"connected"`
}

type Like struct {
	Type     string `json:"type"`
	Likes    string `json:"likes"`
	ID       string `json:"id"`
	HasLikes bool   `json:"hasLikes"`
}

type LikeSent struct {
	ArtPieceID string `json:"artPieceID"`
	UserID     string `json:"userID"`
	ArtShowID  string `json:"artShowID"`
	AddedLike  bool   `json:"addedLike"`//if true it means that the user added a like, otherwise 
}

func serveWs(pool *Pool, w http.ResponseWriter, r *http.Request) {
	fmt.Println("WebSocket Endpoint Hit")
	//upgrade the connection of the
	conn, err := Upgrade(w, r)
	if err != nil {
		fmt.Fprintf(w, "%+v\n", err)
	}

	client := &Client{
		Conn: conn,
		Pool: pool,
	}

	pool.Connect <- client
	client.Read()
}

func setupRoutes() {
	pool := NewPool()
	go pool.Start()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(pool, w, r)
	})
}

func main() {
	fmt.Println("Distributed Chat App v0.01")
	setupRoutes()
	http.ListenAndServe(":8080", nil)
}
