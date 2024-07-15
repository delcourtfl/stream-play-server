package main

import (
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	ip   string // IP address for WebSocket connection
	port string // Port number for WebSocket connection
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
var clients = struct {
	sync.Mutex
	list map[*websocket.Conn]bool
}{list: make(map[*websocket.Conn]bool)}

/**
 * main is the main entry point of the signaling server program. It initializes the WebSocket upgrader,
 * handles WebSocket connections, and starts the HTTP server to listen for incoming connections.
 */
func main() {
	// Get IP and PORT from the args
	args := os.Args[1:] // Exclude the first argument, which is the program name
	if len(args) < 2 {
		log.Println("Please provide both IP address and port as arguments")
		return
	}

	ip = args[0]
	port = args[1]
	log.Println(ip)
	log.Println(port)

	// Handle WebSocket connections
	http.HandleFunc("/stream", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Failed to upgrade WebSocket connection:", err)
			return
		}
		log.Println("Got a new client : " + conn.RemoteAddr().String())
		// Add the new client to the list
		clients.Lock()
		clients.list[conn] = true
		clients.Unlock()
		// Handle signaling messages from the client
		go handleSignaling(conn)
	})

	log.Println("Signaling server started on " + ip + ":" + port)

	go printClientAddresses()

	// Start the HTTP server
	err := http.ListenAndServe(ip+":"+port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

}

func printClientAddresses() {
	for {
		var addresses []string
		for conn := range clients.list {
			addresses = append(addresses, conn.RemoteAddr().String())
		}
		log.Println("Client addresses:", addresses)
		time.Sleep(10 * time.Second) // Adjust the sleep duration as needed
	}
}

/**
 * handleSignaling handles signaling messages from a client WebSocket connection.
 * It continuously reads and broadcasts signaling messages to all connected clients.
 *
 * @param conn A pointer to the websocket.Conn representing the client connection.
 */
func handleSignaling(conn *websocket.Conn) {
	for {
		// Read the signaling message from the client
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("Failed to read signaling message:", err)
			break
		}

		log.Println(string(msg))

		// Broadcast the signaling message to all connected clients
		broadcastMessage(conn, msg)
	}

	// Remove the client from the list when the connection is closed
	clients.Lock()
	delete(clients.list, conn)
	clients.Unlock()

	// Close the WebSocket connection
	err := conn.Close()
	if err != nil {
		log.Println("Failed to close WebSocket connection:", err)
	}
}

/**
 * broadcastMessage broadcasts a signaling message to all connected clients except the sender.
 *
 * @param sender A pointer to the websocket.Conn representing the sender of the message.
 * @param msg A byte slice containing the signaling message to be broadcasted.
 */
func broadcastMessage(sender *websocket.Conn, msg []byte) {
	clients.Lock()

	for client := range clients.list {
		if client == sender {
			continue
		}
		err := client.WriteMessage(websocket.TextMessage, msg)
		// log.Println("Receiver:", client.RemoteAddr().String())
		if err != nil {
			log.Println("Failed to send signaling message to client:", err)
		}
	}

	clients.Unlock()
}
