package main

import (
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

type Client struct {
	IDShowAttending int          //the id of the art show the user is attending
	ID            string          //unique id that identifies a connection
	Conn          *websocket.Conn //websocket connection
	Pool          *Pool           //pool of connections (a client will be in a pool which on our case will represent an art show)
}

//Read method for client, it will read anything the client send to the server
//for now we are just gonna print it and ignore it since we dont need the client to send anything
//(the pointer before the Client means that this method will alter the values of the struct)
func (c *Client) Read() {
	//defer will run this function right before the function returns
	//since it's a infinite loop it will run only if there is an error reading the message
	//that's why we are disconnecting it and closing the connection
	defer func() {
		//it will send through the Disconnect channel the client that is being removed
		//and close the connection
		c.Pool.Disconnect <- c
		c.Conn.Close()
	}()

	for {
		//read the message from the client
		_, p, err := c.Conn.ReadMessage()
		//check for errors
		if err != nil {
			log.Println(err)
			return
		}
		//print the message, we dont really need the user to send anything
		//so we can just print it and ignore it for now
		fmt.Printf("the client %s sent a message: %s\n", c.ID, string(p))
		// c.Pool.LikesHandler <- like
	}
}
