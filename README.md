# Stream Play Server (SPS)
### Version 0.0.2

![SPS Logo](res/logoSPS.png)

Stream Play Server (SPS) is a WebRTC-powered media server for real-time video streaming and remote control of Windows applications enabling Remote Gaming in a simple web browser environment.

## Demo

Small video showing the usage of SPS using a phone with one controller connected to play a game ([Biped](https://store.steampowered.com/app/1071870/Biped/)) remotely

![Demo with Biped game](res/DemoSPS.gif)

## Features

- Low latency video streaming of a window application with WebRTC
- Input catching from a browser instance to the server for remote interaction (using Gamepad API + ViGEm for controllers or Windows API SendInput for mouse and keyboard)

## Description

### Release 0.0.2
### Release notes :

- Upgraded the client UI (simplified overall process)
- Removed `pion WebRTC` and `FFMPEG` dependencies by using the `getDisplayMedia()` method of the MediaDevices interface implemented in (most) browsers.
    
    This was done to improve the performance of the video/audio transmission as I was not able to make it work properly for my use-case using the previous implementation. I would like to avoid depending on the browser methods to focus on a Golang media-server for recording and transmission. I will try to follow the development of the Pion MediaDevice implementation for a future release.

- getDisplayMedia() is browser dependent (https://developer.mozilla.org/en-US/docs/Web/API/MediaDevices/getDisplayMedia) and it was mainly tested on Google Chrome (which enable recording for browser tab video+audio, application video or system video+audio)
- For now, hosting is only available on localhost:web\_port due to the required secure contexts of getDisplayMedia().
- Added Gamepad index modification to be able to separate the multiple clients.

![SPS Diagram 0.0.2](res/NewStructureSPS.png)

### Release 0.0.1
![SPS Diagram 0.0.1](res/OldStructureSPS.png)

- **Media-Server**

Handles real-time communication and peer inputs to the application (Media Processing and Input Management)

- **Signaling Server**

Intermediary for WebRTC clients to exchange session information and coordinate the establishment of direct peer-to-peer connections

- **Client (Webserver)**

User-facing part of the WebRTC application that runs in the web browser (User Interface), 3 tabs available for the connection setup, media stream and controllers setup.

| Connection         | Game         | Controllers         |
| --------------- | --------------- | --------------- |
| ![Connection](res/connection-tab.png) | ![Game](res/game-tab.png) | ![Controllers](res/controllers-tab.png) |

## Custom Installation

Prerequisites : 
- Golang (https://go.dev/doc/install)

Installation steps :
1. Clone the repository.
```
git clone https://github.com/delcourtfl/stream-play-server.git
```
2. Navigate to the project directory.
```cmd
cd stream-play-server
```
3. Install the dependencies before running (optionnal).
```cmd
go get ./...
```
4. Modify the config.json file to change the ip address and ports used.
```json
{
    "ip_address": "192.168.68.101",
    "web_port": "3000",
    "sign_port": "3001",
    "input_port": "3002"
}
```
5. (optional) Modify the media capture settings to suit your needs (webrtc.js).
```js
const displayMediaOptions = { 
    video: getVideo ? {
        width: { ideal: 1280 },
        height: { ideal: 720 },
        frameRate: { max: 30 },
        latency: 0
    } : false,
    audio: getAudio ? {
        noiseSuppression: false,
        autoGainControl: false,
        echoCancellation: false,
        sampleRate: 48000,
        latency: 0
    } : false
};
```
6. Launch the SPS application.
```cmd
go run .
```

## Usage

Once the application is running some commands are available :

- *exit* : close everything
- *stop* : stop the 3 subprocesses (signaling, client, server)
- *sign* : restart signaling server
- *client* : restart client webserver
- *server* : restart media server

When the signaling server is up, you can open the user interface as host (on a local instance only, http://localhost:web_port) or as client (on the ip address you provided in the config.json file, http://ip_address:web_port).

There you can find 3 tabs for managing the game sharing process:
- *Connection Tab*
    - Host button to capture the media and share it using WebRTC
    - Join button to create a P2P connection and read the recorded media
- *Game Tab*
    - Game streamed by the WebRTC connection
- *Controllers Tab*
    - Settings of the controllers which will send inputs back to the host

Enjoy !

## Technologies Used

- HTML, CSS, and JavaScript (for the browser client/host instance).
- Golang for the input management, webserver and signaling server.
- [WebRTC](https://webrtc.org/) for low latency media transmission.
- [ViGEm](https://github.com/ViGEm/ViGEmBus) for game controller emulation.
- [GamepadAPI](https://developer.mozilla.org/en-US/docs/Web/API/Gamepad_API) for input reading.
- [getDisplayMedia](https://developer.mozilla.org/en-US/docs/Web/API/MediaDevices/getDisplayMedia) for game recording.

## Areas of Improvement

This project is a work in progress and as such there are areas that are still being refined and improved. This repository is open to any contributions, suggestions and recommendations for improvements and fixes.

Current issues :
- Video + inputs introduce latency that is still too high for many games (WiFi tests on my network showed a delay of about 0.5 to 1 second between action and visual feedback on 720p 30FPS)
- Audio recording for specific application is not supported in getDisplayMedia() (or not at all for some browsers)
- Controllers are hard to map manually in the client browser
- Gamepad API will need secure context in the future
- Issue where a process won't stop after closing the main golang script
- Automated benchmarking and realtime evaluation could be great

## License

[MIT License](LICENSE)

## Acknowledgement

- The Pion Webrtc implementation and examples (https://github.com/pion/webrtc).
- The excellent cloud-morph application (https://github.com/giongto35/cloud-morph), which was a great starting point for this (smaller) project.
- The stadiacontroller Xbox emulator for ViGEm usage in golang (https://github.com/71/stadiacontroller).
- The tutorial for the visual gamepad interface (https://github.com/CodingWith-Adam/gamepad-tester-simple-just-controller).
- The WebRTC Web demos and samples (https://github.com/webrtc/samples).
- The Pion MediaDevices implementation (for future follow-up) (https://github.com/pion/mediadevices).
