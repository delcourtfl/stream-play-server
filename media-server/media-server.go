package main

import (
	// "context"
	"encoding/json"
	"log"
	"net/url"
	"os"
	"os/signal"
	// "net"
	"time"
	"fmt"

	"io"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
	"github.com/pion/webrtc/v3/pkg/media/h264reader"

	// "github.com/pion/mediadevices"
	// "github.com/pion/mediadevices/pkg/frame"
	// "github.com/pion/mediadevices/pkg/prop"

	// "github.com/pion/mediadevices/pkg/codec/openh264"
)

const (
	h264FrameDuration = time.Millisecond * 333 //33
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
	// captureStream(title)
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
	var clientsVideoTracks []*webrtc.TrackLocalStaticSample
	var clientsAudioTracks []*webrtc.TrackLocalStaticSample
	var clientsConn []*webrtc.PeerConnection

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
func peerConnectionSetup (conn *websocket.Conn, peerConnection *webrtc.PeerConnection) (*webrtc.TrackLocalStaticSample, *webrtc.TrackLocalStaticSample, error) {

	// iceConnectedCtx, _ := context.WithCancel(context.Background())

	// Create a video track
	videoTrack, videoTrackErr := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264}, "video", "pion")
	if videoTrackErr != nil {
		panic(videoTrackErr)
	}

	rtpSender, videoTrackErr := peerConnection.AddTrack(videoTrack)
	if videoTrackErr != nil {
		panic(videoTrackErr)
	}

	go func() {
		rtcpBuf := make([]byte, 1500)
		for {
			if _, _, rtcpErr := rtpSender.Read(rtcpBuf); rtcpErr != nil {
				return
			}
		}
	}()

	go func() {
		dataPipe, err := newCaptureStream(title)

		if err != nil {
			panic(err)
		}

		h264, h264Err := h264reader.NewReader(dataPipe)
		if h264Err != nil {
			panic(h264Err)
		}

		fmt.Println("Try to record")

		// Wait for connection established
		// <-iceConnectedCtx.Done()

		// fmt.Println("BITE2")

		// Send our video file frame at a time. Pace our sending so we send it at the same speed it should be played back as.
		// This isn't required since the video is timestamped, but we will such much higher loss if we send all at once.
		//
		// It is important to use a time.Ticker instead of time.Sleep because
		// * avoids accumulating skew, just calling time.Sleep didn't compensate for the time spent parsing the data
		// * works around latency issues with Sleep (see https://github.com/golang/go/issues/44343)
		// spsAndPpsCache := []byte{}
		// ticker := time.NewTicker(h264FrameDuration)
		// for ; true; <-ticker.C {
		// 	nal, h264Err := h264.NextNAL()
		// 	if h264Err == io.EOF {
		// 		fmt.Printf("All video frames parsed and sent")
		// 		os.Exit(0)
		// 	}
		// 	if h264Err != nil {
		// 		panic(h264Err)
		// 	}

		// 	nal.Data = append([]byte{0x00, 0x00, 0x00, 0x01}, nal.Data...)

		// 	if nal.UnitType == h264reader.NalUnitTypeSPS || nal.UnitType == h264reader.NalUnitTypePPS {
		// 		spsAndPpsCache = append(spsAndPpsCache, nal.Data...)
		// 		continue
		// 	} else if nal.UnitType == h264reader.NalUnitTypeCodedSliceIdr {
		// 		nal.Data = append(spsAndPpsCache, nal.Data...)
		// 		spsAndPpsCache = []byte{}
		// 	}

		// 	fmt.Println("GOGOGOGO")

		// 	if h264Err = videoTrack.WriteSample(media.Sample{Data: nal.Data, Duration: time.Second}); h264Err != nil {
		// 		panic(h264Err)
		// 	}
		// }
		ticker := time.NewTicker(time.Millisecond * 33))
		for ; true; <-ticker.C {
			nal, h264Err := h264.NextNAL()
			if h264Err == io.EOF {
				fmt.Printf("All video frames parsed and sent")
				os.Exit(0)
			}
			if h264Err != nil {
				panic(h264Err)
			}

			if ivfErr != nil {
				panic(ivfErr)
			}

			if h264Err = videoTrack.WriteSample(media.Sample{Data: nal.Data, Duration: time.Second}); h264Err != nil {
				panic(h264Err)
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

	return videoTrack, nil, nil
}
