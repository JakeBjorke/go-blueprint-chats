package main

import (
	"log"
	"net/http"
	"trace"

	"github.com/gorilla/websocket"
)

type room struct {
	//holds incoming messages to be sent to clients
	forward chan []byte
	//channel for clients wanting to join the room
	join chan *client
	//channel for clients wanting to leave the room
	leave chan *client
	//holds all current clients
	clients map[*client]bool
	//tracer will receive trace information of activity om the room
	tracer trace.Tracer
}

//newRoom makes a new room that is ready to go
func newRoom() *room {
	return &room{
		forward: make(chan []byte),
		join:    make(chan *client),
		leave:   make(chan *client),
		clients: make(map[*client]bool),
		tracer:  trace.Off(),
	}
}

func (r *room) run() {
	for {
		select {
		case client := <-r.join:
			//joining
			r.clients[client] = true
			r.tracer.Trace("New client joined")
		case client := <-r.leave:
			//leaving
			delete(r.clients, client)
			close(client.send)
			r.tracer.Trace("Client left")
		case msg := <-r.forward:
			//forward message to all clients
			for client := range r.clients {
				select {
				case client.send <- msg:
					//send the message
					r.tracer.Trace("-- sent to the client")
				default:
					//failed to send
					delete(r.clients, client)
					close(client.send)
					r.tracer.Trace("-- failed to send, cleaned up client")
				}
			}
		}
	}
}

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

var upgrader = &websocket.Upgrader{ReadBufferSize: socketBufferSize, WriteBufferSize: socketBufferSize}

func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("ServeHTTP:  ", err)
		return
	}

	client := &client{
		socket: socket,
		send:   make(chan []byte, messageBufferSize),
		room:   r,
	}

	r.join <- client
	defer func() { r.leave <- client }()
	go client.write()
	client.read()
}
