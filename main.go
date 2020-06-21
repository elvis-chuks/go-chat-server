package main

import (
	"fmt"
	"log"
	"net/http"

	// "github.com/gorilla/websocket"
	"golang.org/x/net/websocket"
)

type Message struct {
	Text string `json:"text"`
	PosX int `json:"posX"`
	PosY int `json:"posY"`
}

type hub struct {
	clients          map[string]*websocket.Conn
	addClientChan    chan *websocket.Conn
	removeClientChan chan *websocket.Conn
	broadcastChan    chan Message
}

var h = newHub()

func handler(ws *websocket.Conn, h *hub) {
	go h.run()
	h.addClientChan <- ws
	log.Println(h)
	for {
	  var m Message
	  err := websocket.JSON.Receive(ws, &m)
	  log.Println(m,"message")
	  if err != nil {
		h.broadcastChan <- Message{err.Error(),m.PosX,m.PosY}
		h.removeClient(ws)
		return
	  }
	  h.broadcastChan <- m
	}
   }

func newHub() *hub {
	return &hub{
		clients:          make(map[string]*websocket.Conn),
		addClientChan:    make(chan *websocket.Conn),
		removeClientChan: make(chan *websocket.Conn),
		broadcastChan:    make(chan Message),
	}
}

func (h *hub) run() {
	for {
		select {
			case conn := <-h.addClientChan:
				h.addClient(conn)
			case conn := <-h.removeClientChan:
				h.removeClient(conn)
			case m := <-h.broadcastChan:
				h.broadcastMessage(m)
		}
	}
	}

func (h *hub) removeClient(conn *websocket.Conn) {
	delete(h.clients, conn.LocalAddr().String())
}
func (h *hub) addClient(conn *websocket.Conn) {
	h.clients[conn.RemoteAddr().String()] = conn
}
   
func (h *hub) broadcastMessage(m Message) {
	for _, conn := range h.clients {
		err := websocket.JSON.Send(conn, m)
		if err != nil {
			fmt.Println("Error broadcasting message: ", err)
		return
		}
	}
}

var clients = 0


func Home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Home page")
}


func setupRoutes(){
	http.HandleFunc("/", Home)
	http.Handle("/ws", websocket.Handler(func(ws *websocket.Conn){
		handler(ws, h)
	}))
}

func main(){
	log.Println("Attempting to start server")
	setupRoutes()

	log.Fatal(http.ListenAndServe(":8080",nil))
}