// Utils function to manage WebRTC connection

import {
    setInputChannel,
    initGamepadInput
} from './inputs.js'; // Adjust the path as needed

let streamIn = null;
let streamOut = null;

let peerConnections = {};
let signalingSocket = null;
let localSocket = null;

let getAudio = true;
let getVideo = true;

function uuidv4() {
    return "10000000-1000-4000-8000-100000000000".replace(/[018]/g, c =>
        (+c ^ crypto.getRandomValues(new Uint8Array(1))[0] & 15 >> +c / 4).toString(16)
    );
};

export let localPeerId = uuidv4();
console.log(localPeerId);

export function getDataChannel() {
    console.log(peerConnections);
    for (const peerId in peerConnections) {
        const peerConnection = peerConnections[peerId];
        const channel = peerConnection.dataChannel;
        if (channel) {
            return channel;
        }
    }
    console.log("No data channel found");
    return null;
}

const processMsg = (msg) => {
    const peerId = msg.peerId;

    switch (msg.type) {
        case 'offer':
            console.log('Received offer');
            if (streamOut) {
                handleOffer(peerId, msg);
            }
            break;
        case 'answer':
            console.log("Received answer");
            if (peerId === localPeerId || !peerConnections[peerId]) {
                console.log('Wrong anwser, ignoring : ', peerId);
            } else if (peerConnections[peerId].peerConnection && peerConnections[peerId].peerConnection.currentRemoteDescription) {
                console.log('No need for answer, ignoring : ', peerId);
            } else {
                handleAnswer(peerId, msg);
            }
            break;
        case 'candidate':
            console.log('Received ICE candidate');
            if (peerConnections[peerId]) {
                handleCandidate(peerId, msg);
            }
            break;
        case 'ready':
            console.log('Received Ready');
            console.log(peerConnections);
            if (peerId === localPeerId || peerConnections[peerId]) {
                console.log('Already in call, ignoring : ', peerId);
            } else {
                makeCall(peerId);
            }
            break;
        case 'join':
            console.log('Received Join');
            console.log(peerConnections);
            if (streamOut) {
                console.log("Sending stream");
                signalingSocket.send(JSON.stringify({
                    type: 'ready',
                    peerId: localPeerId
                }));
            }
            break;
        case 'bye':
            console.log('Received Bye');
            if (peerConnections[peerId]) {
                hangup(peerId);
            }
            break;
        default:
            console.log('Received Unknown message type:', msg.type);
            break;
    }
};

export async function connectToSignServer() {
    try {
        console.log("Loading IP/Port")
        const response = await fetch('wconfig.json');
        const config = await response.json();

        // Access values from the config object
        const serverIP = config.ip_address;
        const serverPort = config.sign_port;
        const inputPort = config.input_port;
        if (signalingSocket) {
            signalingSocket.close();
            signalingSocket = null; // Clean up the reference
            peerConnections = {};
        }
        signalingSocket = new WebSocket('ws://' + serverIP + ':' + serverPort + '/stream');

        signalingSocket.onopen = () => console.log(WebSocket.OPEN);
        signalingSocket.onclose = () => console.log(WebSocket.CLOSED);
        signalingSocket.onerror = () => console.log("Can't connect to signaling server");

        console.log("Server IP:", serverIP, "Server Port:", serverPort);

        signalingSocket.onmessage = async (event) => {
            const message = JSON.parse(event.data);
            console.log(event);
            processMsg(message);
        };

        if (location.hostname === 'localhost') {
            // Is local instance
            localSocket = new WebSocket('ws://localhost:'+inputPort+'/ws');
            localSocket.onopen = () => {
                console.log('Local WebSocket connection opened');
            };
            localSocket.onmessage = (event) => {
                console.log(event);
            }
        }

    } catch (error) {
        console.error('Error loading configuration:', error);
    }
};

async function createPeerConnection(peerId, withDataChannel) {
    if (peerConnections[peerId]) {
        return;
    }
    const peerConnection = new RTCPeerConnection();

    peerConnection.onicecandidate = (e) => {
        const message = {
            type: 'candidate',
            peerId: localPeerId,
            candidate: e.candidate ? e.candidate.candidate : null,
            sdpMid: e.candidate ? e.candidate.sdpMid : null,
            sdpMLineIndex: e.candidate ? e.candidate.sdpMLineIndex : null
        };
        signalingSocket.send(JSON.stringify(message));
    };

    peerConnection.ontrack = (event) => {
        console.log("Got something from", peerId);
        console.log("New track: " + event.track.kind);
        console.log(event.streams);
        const videoElement = document.getElementById('streamVideo');
        videoElement.srcObject = event.streams[0];
        streamIn = event.streams[0];
    };

    if (withDataChannel) {

        const dataChannel = peerConnection.createDataChannel("data");
        dataChannel.onopen = () => {
            console.log("Data channel initiated with", peerId);
            const channel = getDataChannel();
            setInputChannel(channel);
        };
        dataChannel.onmessage = (event) => {
            console.log("Received message from", peerId, ":", event.data);
        };
        peerConnections[peerId] = {
            peerConnection,
            dataChannel,
        };
    
    } else {

        peerConnections[peerId] = {
            peerConnection,
            dataChannel: null
        };

        peerConnection.ondatachannel = (event) => {
            const dataChannel = event.channel;
            console.log(dataChannel);
        
            dataChannel.onopen = () => {
                console.log("Data channel initiated with", peerId);
            };
            dataChannel.onmessage = (event) => {
                console.log("Received input from", peerId);
                if (localSocket && localSocket.readyState === WebSocket.OPEN) {
                    localSocket.send(event.data);
                }
            };

            peerConnections[peerId] = {
                peerConnection,
                dataChannel,
            };
        };

    }
};

async function handleOffer(peerId, offer) {
    await createPeerConnection(peerId, false);
    const peerConnection = peerConnections[peerId].peerConnection;

    if (streamOut) {
        streamOut.getTracks().forEach(track => {
            console.log("Track:", track);
            peerConnection.addTrack(track, streamOut);
        });
    }
    await peerConnection.setRemoteDescription(offer);

    const answer = await peerConnection.createAnswer();
    signalingSocket.send(JSON.stringify({
        type: 'answer',
        peerId: localPeerId,
        sdp: answer.sdp
    }));
    await peerConnection.setLocalDescription(answer);
};


async function handleCandidate(peerId, candidate) {
    const peerConnection = peerConnections[peerId].peerConnection;
    if (!peerConnection) {
        console.error('No peer connection for candidate from', peerId);
        return;
    }
    await peerConnection.addIceCandidate(candidate.candidate ? candidate : null);
};

async function handleAnswer(peerId, answer) {
    const peerConnection = peerConnections[peerId].peerConnection;
    console.log(peerConnection)
    if (!peerConnection) {
        console.error('No peer connection for answer from', peerId);
        return;
    }
    await peerConnection.setRemoteDescription(answer);
};

// Function to check video reception status
export async function checkVideoReceptionStatus() {
    Object.values(peerConnections).forEach((connection) => {
        const peerConnection = connection.peerConnection;
        if (peerConnection) {
            peerConnection.getStats().then((stats) => {
                // printRTCStatsReport(stats);
                stats.forEach((report) => {
                    if (report.type === 'inbound-rtp') {
                        if (report.kind === 'video') {
                            console.log('Video data received: ' + report.framesReceived + ' frames');
                        } else if (report.kind === 'audio') {
                            console.log('Audio data received: ' + report.packetsReceived + ' packets');
                        }
                    }
                });
            });
        }
    });
};

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
    console.log(reportEntries);
};

/////////////////////////////////////////////////////////

export async function hangup(peerId) {
    const peerConnection = peerConnections[peerId].peerConnection;
    if (peerConnection) {
        peerConnection.close();
        delete peerConnections[peerId];
    }
};

export async function hangupCall() {
    for (const peerId in peerConnections) {
        hangup(peerId);
    }
    signalingSocket.send(JSON.stringify({
        type: 'bye',
        peerId: localPeerId
    }));
};

export async function makeCall(peerId) {
    await createPeerConnection(peerId, true);
    const peerConnection = peerConnections[peerId].peerConnection;

    if (streamOut) {
        streamOut.getTracks().forEach(track => {
            console.log("Track:", track);
            peerConnection.addTrack(track, streamOut);
        });
    }

    const offerOptions = {
        offerToReceiveAudio: getAudio ? 1 : 0,
        offerToReceiveVideo: getVideo ? 1 : 0
    };

    const offer = await peerConnection.createOffer(offerOptions);
    signalingSocket.send(JSON.stringify({
        type: 'offer',
        peerId: localPeerId,
        sdp:  offer.sdp
    }));
    await peerConnection.setLocalDescription(offer);
};

export async function startCall() {
    const displayMediaOptions = { video: getVideo, audio: getAudio };

    const stream = await startCapture(displayMediaOptions);
    if (stream) {
        console.log(stream);
        const videoElement = document.getElementById('streamVideo');
        videoElement.srcObject = stream;
        streamOut = stream;

        signalingSocket.send(JSON.stringify({
            type: 'ready',
            peerId: localPeerId
        }));
    } else {
        console.error("Failed to capture media stream.");
    }
};

export async function joinCall() {
    signalingSocket.send(JSON.stringify({
        type: 'join',
        peerId: localPeerId
    }));
}

async function startCapture(displayMediaOptions) {
    try {
        if (!navigator.mediaDevices) {
            console.log("Current instance cannot record media");
            return null
        }
        return await navigator.mediaDevices.getDisplayMedia(displayMediaOptions);
    } catch (err) {
        console.error(err);
        return null;
    }
};

//////////////////////////////////////////////

// Record option Audio / Video

export function toggleAudio () {
    getAudio = !getAudio;
    const audioButton = document.getElementById("recordAudio");
    audioButton.innerText = "Audio : " + (getAudio ? 'On' : 'Off');
    console.log("Audio : " + (getAudio? 'On' : 'Off'));
}

export function toggleVideo () {
    getVideo = !getVideo;
    const videoButton = document.getElementById("recordVideo");
    videoButton.innerText = "Video : " + (getVideo ? 'On' : 'Off');
    console.log("Video : " + (getVideo? 'On' : 'Off'));
}