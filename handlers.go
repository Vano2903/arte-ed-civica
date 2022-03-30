package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	resp "github.com/Vano2903/mostra/responser"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
)

type Handler struct {
	pool     *Pool
	session  *sessions.CookieStore
	upgrader *websocket.Upgrader
}

//just used to log the user interactions with the server
func (h *Handler) loggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		next.ServeHTTP(w, r)
	})
}

//returns all the art shows that are currently exhibiting to the users
func (h *Handler) GetArtShowsForNormalUsers(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: GetArtShows")

	var artShows []ArtShow
	for _, show := range h.pool.ArtShows {
		if show.IsExhibiting {
			//reduce the amount of data sent to the client
			show.ConnectedUsers = len(show.Connected)
			show.Pieces = nil
			artShows = append(artShows, show)
		}
	}

	//send the message to the client
	resp.SuccessJsonParser(w, http.StatusOK, "ArtShows", artShows)
}

//returns all the art shows
func (h *Handler) GetArtShowsForAdmins(w http.ResponseWriter, r *http.Request) {
	conn, err := connectToDB()
	if err != nil {
		resp.Errorf(w, http.StatusInternalServerError, "Error connecting to database: %v", err)
		return
	}
	defer conn.Close()

	shows, err := GetArtShows(conn, false)
	if err != nil {
		resp.Errorf(w, http.StatusInternalServerError, "Error getting art shows: %v", err)
		return
	}

	for i := range shows {
		shows[i].ConnectedUsers = len(shows[i].Connected)
	}
	//send the message to the client
	resp.SuccessJsonParser(w, http.StatusOK, "ArtShows", shows)
}

func (h *Handler) ToggleShow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	//get the id of the artshow
	showName, err := strconv.Atoi(vars["id"])
	if err != nil {
		resp.Errorf(w, http.StatusBadRequest, "Error converting id to int: %v", err)
		return
	}

	show, ok := h.pool.ArtShows[showName]
	if !ok {
		resp.Errorf(w, http.StatusNotFound, "Art show not found: %d", showName)
		return
	}

	if show.IsExhibiting {
		fmt.Println("exhibition is stopping")
		//remove all the clients from the pool which is the artshow
		for _, client := range show.Connected {
			fmt.Printf("the client %s is being disconnected\n", client.ID)
			client.Conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"disconnecting, the artshow is being closed"}`))
			client.Conn.Close()
		}
		//set the art exhibition to false
		show.IsExhibiting = false
		h.pool.ArtShows[showName] = show
		resp.Successf(w, http.StatusOK, "Exhibition %s is being closed by an admin", show.Name)
		return
	}

	//set the artshow itself to true
	show.IsExhibiting = true
	h.pool.ArtShows[showName] = show
	resp.Successf(w, http.StatusOK, "Exhibition %s is being opened by an admin", show.Name)
}

func (h *Handler) GetArtShow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	//get the id of the artshow
	showID, err := strconv.Atoi(vars["id"])
	if err != nil {
		resp.Errorf(w, http.StatusBadRequest, "Error converting id to int: %v", err)
		return
	}

	show, ok := h.pool.ArtShows[showID]
	if !ok {
		resp.Errorf(w, http.StatusNotFound, "Art show not found: %d", showID)
		return
	}

	if !show.IsExhibiting {
		resp.Error(w, http.StatusNotFound, "Art show is not currently exhibiting")
		return
	}

	//create a new session for the user
	session, err := h.session.New(r, "arte-session")
	if err != nil {
		resp.Errorf(w, http.StatusInternalServerError, "Error creating session: %v", err)
		return
	}

	session.Options = &sessions.Options{
		Path: "/",
		//!important
		//the maxage should be set to the duration of the art show and not 7 days
		//for now, for semplicity, we will set it to 7 days but in a final implementation
		//this would be the duration of the art show
		MaxAge: 86400 * 7, //7 days
	}

	conn, err := connectToDB()
	if err != nil {
		resp.Errorf(w, http.StatusInternalServerError, "Error connecting to database: %v", err)
		return
	}
	defer conn.Close()

	id, err := generateNewUser(conn)
	if err != nil {
		resp.Errorf(w, http.StatusInternalServerError, "Error generating new user: %v", err)
		return
	}

	//with this session we can now track the user and the art show he is connected to
	session.Values["showID"] = showID
	session.Values["userID"] = id
	//this will store the ids of the artpieces the user has liked
	session.Values["likes"] = []int{}
	err = session.Save(r, w)
	if err != nil {
		resp.Errorf(w, http.StatusInternalServerError, "Error saving session: %v", err)
		return
	}
	//send the message to the client
	show.GetPieces(conn, true)
	show.ConnectedUsers = len(show.Connected)
	resp.SuccessJsonParser(w, http.StatusOK, "ArtShow", show)
}

func (h *Handler) ToggleLike(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	//get the id of the artshow
	pieceID, err := strconv.Atoi(vars["id"])
	if err != nil {
		resp.Errorf(w, http.StatusBadRequest, "Error converting id to int: %v", err)
		return
	}

	session, err := h.session.Get(r, "arte-session")
	if err != nil {
		resp.Errorf(w, http.StatusInternalServerError, "Error getting session: %v", err)
		return
	}

	fmt.Println("is session nil", session == nil)
	id, ok := session.Values["showID"].(int)
	if !ok {
		resp.Errorf(w, http.StatusBadRequest, "Error getting showID from session, you probably didn't select an art show")
		return
	}

	show, ok := h.pool.ArtShows[id]
	if !ok {
		resp.Errorf(w, http.StatusNotFound, "Art show not found: %d", session.Values["showID"].(int))
		return
	}

	//get the id of the user
	userID, ok := session.Values["userID"].(int)
	if !ok {
		resp.Errorf(w, http.StatusInternalServerError, "Error getting userID from session: %v", err)
		return
	}

	conn, err := connectToDB()
	if err != nil {
		resp.Errorf(w, http.StatusInternalServerError, "Error connecting to database: %v", err)
		return
	}
	defer conn.Close()

	heDid, err := userHasLiked(conn, userID, show.ID, pieceID)
	if err != nil {
		resp.Errorf(w, http.StatusInternalServerError, "Error checking if user has liked art piece: %v", err)
		return
	}

	if heDid {
		show.RemoveLike(conn, userID, pieceID)
		likes := session.Values["likes"]
		session.Values["likes"] = removeFromSlice(likes.([]int), pieceID)
	} else {
		show.AddLike(conn, userID, pieceID)
		likes := session.Values["likes"]
		session.Values["likes"] = append(likes.([]int), pieceID)
	}

	show.GetPieces(conn, true)
	piece := show.GetArtPiece(pieceID, false, nil)
	//broadcast to all the clients connected to the art show via socket that the user has liked the art piece
	for _, client := range show.Connected {
		client.Conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf(`{"type":"like","id":%d, "likes": %d}`, pieceID, piece.Likes)))
	}
	h.pool.ArtShows[session.Values["showID"].(int)] = show
	session.Save(r, w)
	resp.Successf(w, http.StatusOK, "Like toggled")
}

func (h *Handler) ConnectUserToRealTimeUpdates(w http.ResponseWriter, r *http.Request) {
	fmt.Println("WebSocket Endpoint Hit")

	vars := mux.Vars(r)
	showID, err := strconv.Atoi(vars["id"])
	if err != nil {
		resp.Errorf(w, http.StatusBadRequest, "Error converting id to int: %v", err)
		return
	}

	show, ok := h.pool.ArtShows[showID]
	if !ok {
		resp.Errorf(w, http.StatusNotFound, "Art show not found: %d", showID)
		return
	}

	//upgrade the connection of the
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		resp.Errorf(w, http.StatusInternalServerError, "Error upgrading connection: %v", err)
		return
	}

	client := &Client{
		IDShowAttending: show.ID,
		Conn:            conn,
		Pool:            h.pool,
	}

	h.pool.Connect <- client
	client.Read()
}

func NewHandler() *Handler {
	h := &Handler{
		pool:    NewPool(),
		session: sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY"))),
		upgrader: &websocket.Upgrader{
			//size of the incoming buffer
			ReadBufferSize: 1024,
			//size of the transfer buffer
			WriteBufferSize: 1024,
			//should check the origin of the request, for now we return true
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
	h.session.Options = &sessions.Options{Path: "/"}
	return h
}
