import { useState, useEffect, useRef } from 'react';

const CastHook = () => {

    const [captureStream, setCaptureStream] = useState(null);
    const captureRef = useRef(null);

    const [videoStream, setVideoStream] = useState(null);
    const videoRef = useRef(null);

    const [readyState, setReadyState] = useState(WebSocket.CLOSED);

    const [signalingSocket, setSignalingSocket] = useState(null);
    
    const [peerConnection, setPeerConnection] = useState(null);
    const [dataChannel, setDataChannel] = useState(null);

    const [lastMessage, setLastMessage] = useState(null);

    const [startDisabled, setStartDisabled] = useState(true);
    const [hangupDisabled, setHangupDisabled] = useState(true);

    // Function to fetch and use the configuration
    async function loadConfig() {

        try {
            console.log("Loading IP/Port")
            const response = await fetch('wconfig.json');
            const config = await response.json();

            // Access values from the config object
            const serverIP = config.ip_address;
            const serverPort = config.port;
            const signalingSocket = new WebSocket('ws://' + serverIP + ':' + serverPort + '/stream');
            setReadyState(WebSocket.CONNECTING);
            signalingSocket.onopen = () => setReadyState(WebSocket.OPEN);
            signalingSocket.onclose = () => setReadyState(WebSocket.CLOSED);

            console.log("Server IP:", serverIP, "Server Port:", serverPort);

            signalingSocket.onmessage = async (event) => {
                const message = JSON.parse(event.data);
                setLastMessage(message);
            };

            setSignalingSocket(signalingSocket);
            console.log("Load Config Success");
            setStartDisabled(false);
        } catch (error) {
            console.error('Error loading configuration:', error);
        }
    }

    useEffect(() => {
        // Check the last message whenever it changes
        if (lastMessage) {

            switch (lastMessage.type) {
                case 'offer':
                    console.log('Received offer');
                    handleOffer(lastMessage);
                    break;
                case 'answer':
                    console.log("Received answer");
                    handleAnswer(lastMessage);
                    break;
                case 'candidate':
                    console.log('Received ICE candidate');
                    handleCandidate(lastMessage);
                    break;
                case 'ready':
                    console.log('Received Ready');
                    if (peerConnection) {
                        console.log('Already in call, ignoring');
                    } else {
                        makeCall();
                    }
                    break;
                case 'bye':
                    console.log('Received Bye');
                    if (peerConnection) {
                        hangup();
                    }
                    break;
                default:
                    console.log('Received Unknown message type:', lastMessage.type);
                    break;
            }
        }
    }, [lastMessage]);

    async function handleOffer(offer) {
        if (peerConnection) {
            console.error('existing peerconnection for offer');
            return;
        }
    
        const pc = await createPeerConnection();
        if (captureStream) {
            captureStream.getTracks().forEach(track => {
                console.log("Track:", track);
                pc.addTrack(track, captureStream);
            });
        }
        await pc.setRemoteDescription(offer);
        
        const answer = await pc.createAnswer();
        signalingSocket.send(JSON.stringify({
            type: 'answer',
            sdp:  answer.sdp
        }));
        await pc.setLocalDescription(answer);
        setPeerConnection(pc);
    }
    
    async function handleCandidate(candidate) {
        if (!peerConnection) {
            console.error('no peerconnection for candidate');
            return;
        }
        if (!candidate.candidate) {
            await peerConnection.addIceCandidate(null);
        } else {
            await peerConnection.addIceCandidate(candidate);
        }
    }

    async function handleAnswer(answer) {
        if (!peerConnection) {
            console.error('no peerconnection for answer');
            return;
        }
        await peerConnection.setRemoteDescription(answer);
    }

    async function createPeerConnection() {

        const pc = new RTCPeerConnection();

        pc.onicecandidate = e => {
            const message = {
                type: 'candidate',
                candidate: null,
            };
            if (e.candidate) {
                message.candidate = e.candidate.candidate;
                message.sdpMid = e.candidate.sdpMid;
                message.sdpMLineIndex = e.candidate.sdpMLineIndex;
            }
            signalingSocket.send(JSON.stringify(message));
        };

        pc.ontrack = function (event) {
            console.log("Got something")
            console.log("new track : " + event.track.kind);
            console.log(event.streams[0]);
            setVideoStream(event.streams[0]);
        };

        // Create a data channel to send data
        const dataChannel = pc.createDataChannel("data");

        dataChannel.addEventListener("open", () => {
            console.log("Data channel initiated");
        });

        dataChannel.onmessage = function (event) {
            var message = event.data;
            console.log("Received message:");
            console.log(message);
        };

        setDataChannel(dataChannel);

        return pc;
    }

    async function hangup() {
        if (peerConnection) {
            peerConnection.close();
            setPeerConnection(null);
            // peerConnection = null;
        }
        captureStream.getTracks().forEach(track => track.stop());
        setCaptureStream(null);
        setStartDisabled(false);
        setHangupDisabled(true);
    };
  
    async function makeCall() {
        const pc = await createPeerConnection();

        if (captureStream) {
            captureStream.getTracks().forEach(track => {
                console.log("Track:", track);
                pc.addTrack(track, captureStream);
            });
            // captureStream.getTracks().forEach(track => peerConnection.addTrack(track, captureStream));
        }

        const offerOptions = {
            offerToReceiveAudio: 1,
            offerToReceiveVideo: 1
        };
  
        const offer = await pc.createOffer(offerOptions);
        signalingSocket.send(JSON.stringify({
            type: 'offer',
            sdp:  offer.sdp
        }));
        await pc.setLocalDescription(offer);
        setPeerConnection(pc);
    }
    
    const hangupCall = () => {
        hangup();
        signalingSocket.send(JSON.stringify({
            type: 'bye',
        }));
    };

    // useEffect(() => {
    //     if (videoRef && videoRef.current && videoRef.current.srcObject) {
    //         console.log("media stream object created");
    //         videoRef.current.srcObject = videoStream;
    //     } else {
    //         console.log("Can't add stream to video element ...");
    //         console.log(videoStream);
    //     }
    // }, [videoStream, videoRef]);

    return { 
        // setupDisplay,
        captureRef,
        setCaptureStream,
        captureStream,
        setVideoStream,
        videoStream,
        hangupCall,
        startDisabled,
        setStartDisabled,
        hangupDisabled,
        setHangupDisabled,
        loadConfig,
        signalingSocket,
        videoRef
    };
};

export default CastHook;