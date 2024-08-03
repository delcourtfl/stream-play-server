package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

type KeyEvent struct {
	Type   string `json:"type"`
	Data   string `json:"data"`
	Option int    `json:"option"`
}

type KeyData struct {
	KeyCode int `json:"keyCode"`
}

var (
	// ip    string // IP address for WebSocket connection
	port  string // Port number for WebSocket connection
	title string // Title of the application to capture
)

type WSMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

/**
 * main is the entry point of the webrtc server program.
 * It sets up interrupt handling, starts capturing the specified application window,
 * establishes a WebSocket connection, creates peer connections and video tracks for clients,
 * opens a UDP listener for RTP packets, handles signaling messages, and captures and sends video frames.
 */
func main() {
	// Get Port from the args
	args := os.Args[1:] // Exclude the first argument, which is the program name
	if len(args) < 1 {
		log.Println("Please provide the input port as argument")
		return
	}

	// No ip is needed -> localhost only
	port = args[0]
	log.Println(port)
	addr := "localhost:" + port

	http.HandleFunc("/ws", handleConnection)

	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

}

func processMessage(msg []byte, x360Array *[4]*Xbox360Controller, emulatorArray *[4]*Emulator, title string) {

	var event KeyEvent
	err := json.Unmarshal(msg, &event)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return
	}

	log.Println("Received message  :")
	log.Println(string(msg))

	switch event.Type {

	// Disable keyboard inputs for now (no use and kinda bad)
	// case "KEYDOWN":
	// 	var data KeyData
	// 	err = json.Unmarshal([]byte(event.Data), &data)
	// 	if err != nil {
	// 		fmt.Println("Error parsing KeyData:", err)
	// 		return
	// 	}
	// 	sendKey(uint16(data.KeyCode), true, title)

	// case "KEYUP":
	// 	var data KeyData
	// 	err = json.Unmarshal([]byte(event.Data), &data)
	// 	if err != nil {
	// 		fmt.Println("Error parsing KeyData:", err)
	// 		return
	// 	}
	// 	sendKey(uint16(data.KeyCode), false, title)

	case "GAMEPAD":
		var data xusb_report
		err = json.Unmarshal([]byte(event.Data), &data)
		if err != nil {
			fmt.Println("Error parsing ControllerData:", err)
			return
		}

		report := Xbox360ControllerReport{}
		report.native = data

		if event.Option > len(*x360Array)-1 {
			fmt.Println("Wrong index for gamepad")
			return
		}

		err := (*x360Array)[event.Option].Send(&report)
		if err != nil {
			fmt.Println(err)
		}

	case "ADDGAMEPAD":
		index := event.Option
		fmt.Println("Adding a new gamepad: ", index)

		if index >= 4 {
			fmt.Println("Can't have more than 4 gamepads")
			return
		}

		if x360Array[index] != nil {
			fmt.Println("A gamepad is already set at index", index)
			return
		}

		emulator, err := NewEmulator(nil)
		if err != nil {
			fmt.Printf("unable to start ViGEm client: %v\n", err)
		}

		x360, err := emulator.CreateXbox360Controller()
		if err != nil {
			fmt.Printf("unable to create emulated Xbox 360 controller: %v\n", err)
		}

		if err = x360.Connect(); err != nil {
			fmt.Printf("unable to connect to emulated Xbox 360 controller: %v\n", err)
		}

		x360Array[index] = x360
		emulatorArray[index] = emulator

		time.Sleep(5 * time.Second)
		fmt.Printf("Gamepad OK \n")

	default:
		log.Println("Unknown event type:", event.Type)

	}
}

func handleConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error while upgrading connection:", err)
		return
	}
	var x360Array [4]*Xbox360Controller
	var emulatorArray [4]*Emulator
	defer func() {
		// Cleanup resources when the connection is closed
		for _, emulator := range emulatorArray {
			if emulator != nil {
				emulator.Close()
			}
		}
		for _, x360 := range x360Array {
			if x360 != nil {
				x360.Close()
			}
		}
		conn.Close()
		log.Println("Connection closed and resources cleaned up")
	}()

	log.Println("Client connected")

	for {
		_, msg, _ := conn.ReadMessage()

		var event KeyEvent
		err = json.Unmarshal(msg, &event)
		if err != nil {
			fmt.Println("Error parsing JSON:", err)
			return
		}

		processMessage(msg, &x360Array, &emulatorArray, title)
	}
}
