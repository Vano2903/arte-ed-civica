package main

import (
	"fmt"
	"log"
)

/*
Before reading this code we should talk about what are channels and what are goroutines.

Goroutines:
They are lightweight threads that can be executed in parallel
and have some concurrency and zombie/orphan treads protections.

The main advantage of goroutines is that they are lightweight (way more then threads or subprocesses)
and they are easy to create and dispose.

To create a goroutines we just need to add the keyword 'go' in front of
a function and this will execute in parallel.

To explain better the orphan/zombie thread concept, we must say that goroutines can stay alive
only if they didn't terminate the execution yet and if the parent routine is still running.
The last part, if the parent routine is still running, prevent the runtime to execute zombie threads because
when the execution ends all the routines (even if they are still running) will be automatically terminated

Channels:
They are a way to communicate between goroutines, since goroutines can not return values because
the program would have to wait for the returns of it, which would destroy the purpose of having a
multithreading program, with channels we can communicate between goroutines and thus return values.

Channels have returns type (since everything in go is strongly typed) and they accept only that type
as transfer value.

Basically a channel is a pipe between threads but it's strongly typed.

to send/receive values we use the '<-' operators.
to send we do 'channel <- value'
to receive we do 'var value type = <-channel'
*/

//Pool is a struct that represents a pool of connections
type Pool struct {
	Connect    chan *Client
	Disconnect chan *Client
	Clients    map[*Client]bool
	ArtShows   map[int]ArtShow
}

//this function will listen for incoming websockets connections
//it will be called only by one pool instance because otherwise we would run into multiple
//goroutines listening for the same client and it would create write issues
func (pool *Pool) Start() {
	for {
		//the select keyword will execute a function depending by the channel that is returning something
		select {
		//case where a client is connectng
		case client := <-pool.Connect:
			if !pool.ArtShows[client.IDShowAttending].IsExhibiting {
				//it will respond saying that the artshow is closed and close the connection
				client.Conn.WriteMessage(1, []byte(`{"type":"closed","msg":"the artshow is closed"}`))
				client.Conn.Close()
				break
			}
			//we add the client to the pool
			pool.Clients[client] = true
			//and we also add the client to the art show list
			show := pool.ArtShows[client.IDShowAttending]
			show.AddClient(client)
			pool.ArtShows[client.IDShowAttending] = show
			fmt.Println("Connected Clients: ", len(pool.Clients))
			//and we broadcast that a new client has connected
			for client, _ := range pool.Clients {
				fmt.Println(client)
				client.Conn.WriteMessage(1, []byte(fmt.Sprintf(`{"type":"join","connected":%d}`, len(pool.Clients))))
			}
		//case where a client is disconnecting
		case client := <-pool.Disconnect:
			//deleting a client from the pool (as we defined before the pool is a map o we need the delete keyword)
			delete(pool.Clients, client)
			fmt.Println("Connected Clients:", len(pool.Clients))
			//and broadcast that the user disconnected
			for client, _ := range pool.Clients {
				client.Conn.WriteMessage(1, []byte(fmt.Sprintf(`{"type":"join","connected":%d}`, len(pool.Clients))))
			}

			//case where a message is received
			//!this case is not in use for now
			// case like := <-pool.LikesHandlr
			// fmt.Printf("The connection %s hs liked the picture s, now it has %s likes\n", like.ID, like.ID, like.Likes)
			// // we send the message to all the clients in the pool
			// for client, _ := range pool.Clients {
			// 	if err = client.Conn.WriteJSON(messae); err != nil {
			// 		fmt.Println(err)
			// 		return
			// 	}
			// }
		}
	}
}

//NewPool is like a costructor, usually we dont need a constructor
//but since we are using channels, maps and pointers we need to intantiate them beforehand
func NewPool() *Pool {
	p := &Pool{
		Connect:    make(chan *Client),
		Disconnect: make(chan *Client),
		Clients:    make(map[*Client]bool),
		ArtShows:   make(map[int]ArtShow),
	}

	//add all the art shows to the pools
	conn, _ := connectToDB()
	defer conn.Close()
	shows, err := GetArtShows(conn, false)
	if err != nil {
		log.Fatal("error getting art shows: ", err)
	}
	for _, show := range shows {
		p.ArtShows[show.ID] = show
	}
	return p
}
