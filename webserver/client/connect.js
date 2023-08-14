document.getElementById("connectPC").addEventListener("click", setupPeerConnection);
document.getElementById("retryFFMPEG").addEventListener("click", setupDisplay);
document.getElementById("testAudio").addEventListener("click", testAudio);
document.getElementById("testVideo").addEventListener("click", testVideo);

var testAudioButton = false;
var testVideoButton = true;

let streamSocket, peerConnection, dataChannel;

// Function to fetch and use the configuration
async function loadConfig() {
    try {
        console.log("Loading IP/Port")
        const response = await fetch('wconfig.json');
        const config = await response.json();

        // Access values from the config object
        const serverIP = config.ip_address;
        const serverPort = config.port;
        streamSocket = new WebSocket('ws://' + serverIP + ':' + serverPort + '/stream');

        console.log("Server IP:", serverIP);
        console.log("Server Port:", serverPort);

        peerConnection = new RTCPeerConnection();
        // Create a data channel to send data
        dataChannel = peerConnection.createDataChannel("data");

        dataChannel.addEventListener("open", () => {
            console.log("Data channel initiated");
        });

        dataChannel.onmessage = function(event) {
            var message = event.data;
            console.log("Received message:");
            console.log(message);
        };

        streamSocket.onmessage = function (event) {
            const message = JSON.parse(event.data);
            console.log(message);

            if (message.type === 'offer') {
                console.log("No need for an offer");
            } else if (message.type === 'answer') {
                handleAnswer(message);
            }
        };

    } catch (error) {
        console.error('Error loading configuration:', error);
    }
}

// Call the function to load the configuration
loadConfig();

function testAudio () {
    testAudioButton = !testAudioButton;
    document.getElementById("testAudio").innerText = "Audio : " + testAudioButton;
    console.log("Audio : " + testAudioButton);
}
function testVideo () {
    testVideoButton = !testVideoButton;
    document.getElementById("testVideo").innerText = "Video : " + testVideoButton;
    console.log("Video : " + testVideoButton);
}

function handleAnswer(message) {
    console.log("Got an answer");
    console.log(message.data);

    try {
        peerConnection.setRemoteDescription(new RTCSessionDescription(message.data))
    } catch (e) {
        alert(e)
    }
}

function handleIceCandidate(event) {
    console.log("add ice");
    if (event.candidate) {
        streamSocket.send(JSON.stringify({
            type: 'candidate',
            data: event.candidate
        }));
    }
}

function handleCandidate(event) {
    const message = JSON.parse(event.data);
    console.log(message);

    if (message.type === 'candidate') {
        peerConnection.addIceCandidate(new RTCIceCandidate(message.data))
            .catch(error => {
                console.log("Failed to add ICE candidate:", error);
            });
    }
}

function send(data) {
    dataChannel.send(JSON.stringify(data));
}

function setupPeerConnection() {
    console.log("Sending Offer");

    peerConnection.oniceconnectionstatechange = function(event) {
        console.log(peerConnection.iceConnectionState);
        if (peerConnection.iceConnectionState === 'connected') {
            console.log("Closing WebSocket");
            streamSocket.close();
            // initInput();
        }
    };

    peerConnection.onicecandidate = handleIceCandidate;

    peerConnection.ontrack = function (event) {
        console.log("new track : " + event.track.kind);

        if (!video.srcObject) {
            console.log("media stream object created");
            video.srcObject = event.streams[0];
        } else {
            video.srcObject.addTrack(event.track);
        }
    };

    if (testAudioButton) {
        peerConnection.addTransceiver('audio', {'direction': 'recvonly'})
    } 
    
    if (testVideoButton) {
        peerConnection.addTransceiver('video', {'direction': 'recvonly'})
    }

    peerConnection.createOffer()
        .then(offer => {
            peerConnection.setLocalDescription(offer);
            streamSocket.send(JSON.stringify({
                type: 'offer',
                data:  offer
            }));
        });

    // Function to check video reception status
    function checkVideoReceptionStatus() {
        peerConnection.getStats().then((stats) => {
            printRTCStatsReport(stats);
            stats.forEach((report) => {
                if (report.type === 'inbound-rtp' && report.kind === 'video') {
                    console.log('Video data received: ' + report.framesReceived + ' frames');
                }
            });
        });
    }

    // Function to print RTCStatsReport cleanly using console.table()
    function printRTCStatsReport(report) {
        const reportEntries = [];
    
        report.forEach((value, key) => {
            // Skip entries that are not objects (e.g., report ID)
            if (typeof value !== 'object' || value === null) {
                return;
            }
        
            const entry = {
                Type: value.type,
                ID: key,
            };
        
            // Extract and add properties to the entry
            for (const prop in value) {
                if (prop !== 'type') {
                entry[prop] = value[prop];
                }
            }
    
            reportEntries.push(entry);
        });
    
        // Print the report using console.table()
        // console.table(reportEntries);
        console.log(reportEntries);
    }
    
    // Call the function periodically to monitor video reception
    setInterval(checkVideoReceptionStatus, 10000); // Check every second (adjust as needed)
}

function setupDisplay() {
    console.log("RETRY DISPLAY");
    send({
        type: "DISPLAY",
        data: "",
    });
}