package main

import (
	"encoding/json"
	"log"
	"net/url"
	"os"
	"os/signal"
	"net"
	"time"
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
)

type KeyEvent struct {
	Type string `json:"type"`
	Data string `json:"data"`
	Option int `json:"option"`
}

type KeyData struct {
	KeyCode int `json:"keyCode"`
}

var (
	ip    string 		// IP address for WebSocket connection
	port  string        // Port number for WebSocket connection
	title string 		// Title of the application to capture
)

type WSMessage struct {
	Type string `json:"type"`
	Data json.RawMessage `json:"data"`
}

/**
 * main is the entry point of the webrtc server program.
 * It sets up interrupt handling, starts capturing the specified application window,
 * establishes a WebSocket connection, creates peer connections and video tracks for clients,
 * opens a UDP listener for RTP packets, handles signaling messages, and captures and sends video frames.
 */
func main() {
	// Get IP, PORT and TITLE from the args
	args := os.Args[1:] // Exclude the first argument, which is the program name
	if len(args) < 2 {
		log.Println("Please provide both IP address, port, [exec] as arguments")
		return
	}

	ip = args[0]
	port = args[1]
	if len(args) > 2 {
		title = args[2]
	} else {
		title = ""
	}
	log.Println(ip)
	log.Println(port)
	log.Println(title)

	// Set up interrupt channel to handle interrupt signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// Start capturing the specified application window
	captureStream(title)
	// Use a defer statement to ensure the command process is killed when the main function exits
	defer stopAllCapture()
	log.Println("Capturing...")

	// Establish the WebSocket connection
	u := url.URL{Scheme: "ws", Host: ip + ":" + port, Path: "/stream"}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("Failed to connect to signaling server:", err)
	}
	defer conn.Close()

	// Create a map to store peer connections andd tracks for each client
	var clientsVideoTracks []*webrtc.TrackLocalStaticRTP
	var clientsAudioTracks []*webrtc.TrackLocalStaticRTP
	var clientsConn []*webrtc.PeerConnection

	bufferSize := 300000 // 300KB

	// Open a UDP Listener for RTP Packets on port 5004
	listenerVideo, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 5004})
	if err != nil {
		panic(err)
	}
	err = listenerVideo.SetReadBuffer(bufferSize)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err = listenerVideo.Close(); err != nil {
			panic(err)
		}
	}()

	// Open a UDP Listener for RTP Packets on port 5005
	listenerAudio, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 5005})
	if err != nil {
		panic(err)
	}
	err = listenerAudio.SetReadBuffer(bufferSize)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err = listenerAudio.Close(); err != nil {
			panic(err)
		}
	}()

	done := make(chan struct{})
	message := &WSMessage{}

	// Goroutine to handle signaling messages from the client
	go func() {
		var peerConnection *webrtc.PeerConnection
		for {
			// Read the signaling message from the client
			_, msg, err := conn.ReadMessage()

			if err != nil {
				log.Println("Failed to read signaling message:", err)
				break
			}

			err = json.Unmarshal(msg, &message)
			if err != nil {
				log.Println("Failed to unmarshal signaling message:", err)
				continue
			}

			if message.Type == "offer" {
				log.Println("TRYING OFFER")

				// Create a new PeerConnection for the offer
				peerConnection, err = webrtc.NewPeerConnection(webrtc.Configuration{})
				if err != nil {
					log.Fatal("Failed to create WebRTC peer connection:", err)
				}
				defer peerConnection.Close()

				log.Println(peerConnection)

				clientsConn = append(clientsConn, peerConnection)

				// Set up video track for the client
				newVideoTrack, newAudioTrack, err := peerConnectionSetup(conn, peerConnection)
				clientsVideoTracks = append(clientsVideoTracks, newVideoTrack)
				clientsAudioTracks = append(clientsAudioTracks, newAudioTrack)

				// Unmarshal and set the remote description
				offer := webrtc.SessionDescription{}
				if err := json.Unmarshal(message.Data, &offer); err != nil {
					log.Println("Failed to Unmarshal answer:", err)
					continue
				}
				if err := peerConnection.SetRemoteDescription(offer); err != nil {
					log.Println("Failed to SetRemoteDescription:", err)
					continue
				}

				// Create an answer
				answer, err := peerConnection.CreateAnswer(nil)
				if err != nil {
					log.Println("Failed to create answer:", err)
					continue
				}

				gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

				// Set the local description
				if err := peerConnection.SetLocalDescription(answer); err != nil {
					log.Println("Failed to SetLocalDescription:", err)
					continue
				}
		
				<-gatherComplete

				answerToSend, err := json.Marshal(peerConnection.LocalDescription())
				if err != nil {
					log.Println(err)
					continue
				}

				// Send the answer to the client
				log.Println("Sending answer")
				if err = conn.WriteJSON(&WSMessage{
					Type: "answer",
					Data:  answerToSend,
				}); err != nil {
					log.Println(err)
					continue
				}

			} else if message.Type == "answer" {
				log.Println("No need for answer")

			} else if message.Type == "candidate" {
				log.Println("Trying candidate")
				// Unmarshal and add the ICE candidate
				candidate := webrtc.ICECandidateInit{}
				if err := json.Unmarshal(message.Data, &candidate); err != nil {
					log.Println("Failed to unmarshal candidate:", err)
					continue
				}
				if err := peerConnection.AddICECandidate(candidate); err != nil {
					log.Println("Failed to add ICE candidate:", err)
					continue
				}
				log.Println(candidate)

			} else {
				log.Println("Unknown message type:", message.Type)
			}
		}
	}()

	// Goroutine to handle capturing and sending video frames
	go func() {
		// Read RTP packets forever and send them to the WebRTC Client
		inboundRTPPacket := make([]byte, 1600) // UDP MTU
		for {
			n, _, err := listenerVideo.ReadFrom(inboundRTPPacket)
			if err != nil {
				panic(err)
			}

			// Send RTP packets to all connected clients
			for _, track := range clientsVideoTracks {
				_, err = track.Write(inboundRTPPacket[:n])
				if err != nil {
					panic(err)
				}
			}
		}
	}();

	// Goroutine to handle capturing and sending audio frames
	go func() {
		// Read RTP packets forever and send them to the WebRTC Client
		inboundRTPPacket := make([]byte, 1600) // UDP MTU
		for {
			n, _, err := listenerAudio.ReadFrom(inboundRTPPacket)
			if err != nil {
				panic(err)
			}

			// Send RTP packets to all connected clients
			for _, track := range clientsAudioTracks {
				_, err = track.Write(inboundRTPPacket[:n])
				if err != nil {
					panic(err)
				}
			}
		}
	}();
	
	// Send messages to the server
	for {
		select {
		case <-done:
			return
		case <-interrupt:
			log.Println("Interrupt signal received, closing connection...")
			return
		}
	}

}

/**
 * Sets up a peer connection by adding a video track, handling incoming RTCP packets,
 * creating a data channel to receive data, and setting event handlers for ICE candidates and tracks.
 *
 * @param conn - WebSocket connection
 * @param peerConnection - WebRTC peer connection
 * @return videoTrack - Video track for the client
 * @return error - Error, if any
 */
func peerConnectionSetup (conn *websocket.Conn, peerConnection *webrtc.PeerConnection) (*webrtc.TrackLocalStaticRTP, *webrtc.TrackLocalStaticRTP, error) {

	// Create a video track            
	videoTrack, err := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264}, "video", "pion")
	if err != nil {
		panic(err)
	}

	rtpVideoSender, err := peerConnection.AddTrack(videoTrack)
	if err != nil {
		log.Fatal("Failed to add video track:", err)
	}

	// Read incoming RTCP packets
	// Before these packets are returned they are processed by interceptors. For things
	// like NACK this needs to be called.
	go func() {
		rtcpBuf := make([]byte, 1500)
		for {
			if _, _, rtcpErr := rtpVideoSender.Read(rtcpBuf); rtcpErr != nil {
				log.Println("Error on RTCP reader")
				return
			}
		}
	}()

	// Create an audio track
	audioTrack, err := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus}, "audio", "pion")
	if err != nil {
		panic(err)
	}
	
	rtpAudioSender, err := peerConnection.AddTrack(audioTrack)
	if err != nil {
		log.Fatal("Failed to add audio track:", err)
	}

	go func() {
		rtcpBuf := make([]byte, 1500)
		for {
			if _, _, rtcpErr := rtpAudioSender.Read(rtcpBuf); rtcpErr != nil {
				log.Println("Error on RTCP reader")
				return
			}
		}
	}()

	// Create a data channel to receive data
	peerConnection.OnDataChannel(func(dataChannel *webrtc.DataChannel) {

		var x360Array []*Xbox360Controller
		var emulatorArray []*Emulator

		dataChannel.OnOpen(func() {
			fmt.Println("Data channel initialized")
		})

		dataChannel.OnClose(func() {
			fmt.Println("Data channel closed")
			for _, emulator := range emulatorArray {
				emulator.Close()
			}
			for _, x360 := range x360Array {
				x360.Close()
			}
		})

		dataChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
			log.Println("Received message  :")
			log.Println(string(msg.Data))

			var event KeyEvent
			err := json.Unmarshal([]byte(msg.Data), &event)
			if err != nil {
				fmt.Println("Error parsing JSON:", err)
				return
			}

			switch event.Type {
				case "KEYDOWN":
					var data KeyData
					err = json.Unmarshal([]byte(event.Data), &data)
					if err != nil {
						fmt.Println("Error parsing KeyData:", err)
						return
					}
					sendKey(uint16(data.KeyCode), true, title)

				case "KEYUP":
					var data KeyData
					err = json.Unmarshal([]byte(event.Data), &data)
					if err != nil {
						fmt.Println("Error parsing KeyData:", err)
						return
					}
					sendKey(uint16(data.KeyCode), false, title)

				case "GAMEPAD":
					var data xusb_report
					err = json.Unmarshal([]byte(event.Data), &data)
					if err != nil {
						fmt.Println("Error parsing ControllerData:", err)
						return
					}

					report := Xbox360ControllerReport{}
					report.native = data
					
					if (event.Option > len(x360Array) - 1){
						fmt.Println("Wrong index for gamepad")
						return
					}

					err := x360Array[event.Option].Send(&report)
					if err != nil {
						fmt.Println(err)
					}

				case "ADDGAMEPAD":
					fmt.Println("Adding a new gamepad")
					if (len(x360Array) >= 4){
						fmt.Println("Already 4 gamepads")
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

					emulatorArray = append(emulatorArray, emulator)
					x360Array = append(x360Array, x360)
				
					time.Sleep(5 * time.Second)
					fmt.Printf("Gamepad OK \n")

				case "DISPLAY":
					fmt.Println("Trying to refresh the display")
					restartFFMPEG(title)

			}
		})
	})

	// Set up event handlers for ICE candidates and tracks
	peerConnection.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		log.Println("Adding candidate")
		if candidate != nil {
			conn.WriteJSON(map[string]interface{}{
				"type":     "candidate",
				"data":  	candidate,
			})
		}
	})

	// Set the handler for Peer connection state
	// This will notify you when the peer has connected/disconnected
	peerConnection.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		log.Println("Peer Connection State has changed:")
		log.Println(s)

		if s == webrtc.PeerConnectionStateFailed {
			log.Println("Peer Connection has gone to failed exiting")
		} else if s == webrtc.PeerConnectionStateConnected {
			log.Println("Good to go !")
		}
	})

	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		log.Println("Connection State has changed")
		log.Println(connectionState.String())
		if connectionState == webrtc.ICEConnectionStateFailed {
			if closeErr := peerConnection.Close(); closeErr != nil {
				// panic(closeErr)
				log.Println(closeErr)
			}
		} else if connectionState == webrtc.ICEConnectionStateConnected {
			log.Println("Good, connection was established !")
		}
	})

	return videoTrack, audioTrack, nil
}
