package main

import (
	"log"

	"github.com/gorilla/websocket"
)

type Client struct {
	ID   string          //unique id that identifies a connection
	Conn *websocket.Conn //websocket connection
	Pool *Pool           //pool of connections (a client will be in a pool which on our case will represent an art show)
}

//Read method for client, it will read the like signal, save it into
//the database and broadcast to the pool of clients the changes
//(the pointer before the Client means that this method will alter the values of the struct)
func (c *Client) Read() {
	//defer will run this function right before the function returns
	defer func() {
		//it will send through the Disconnect channel the client that is being removed
		//and close the connection
		c.Pool.Disconnect <- c
		c.Conn.Close()
	}()

	for {
		//read the message from the client
		messageType, p, err := c.Conn.ReadMessage()
		//check for errors
		if err != nil {
			log.Println(err)
			return
		}
		//convert the message to the Message struct and breadcast it to the pool
		like := Like{Type: messageType, Body: string(p)}
		c.Pool.LikesHandler <- like
	}
}
