package main

import (

	"log"
	"net/http"
	"github.com/gorilla/websocket"
)
var clients = make(map[*websocket.Conn]bool) // Map that will track clients
var broadcasts = make(chan Message) // Channel that will hold messages
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Message struct {
	Username	string	`json:"username"`
	Message 	string 	`json:"message"`
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	socket, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer socket.Close()

	clients[socket] = true 
	for {
		var msg Message
		err := socket.ReadJSON(&msg)
		if err != nil {
			log.Printf("error: %v", err)
			delete(clients, socket)
			break
		}
		broadcasts <- msg // send the message to the broadcast channel
	}
}

// Broadcast messages to all clients
func handleMessages() {
	for {
		msg := <-broadcasts
		for client := range clients {
			err := client.WriteJSON(&msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}

func main() {
	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/", fs)

	http.HandleFunc("/ws", handleConnections)

	go handleMessages() // start goroutine to handle incoming chat messages concurrently

	// start server
	log.Println("http server started on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
