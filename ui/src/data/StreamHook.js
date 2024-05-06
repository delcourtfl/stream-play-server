import { useState, useEffect, useRef } from 'react';

const StreamHook = () => {
    const [videoStream, setVideoStream] = useState(null);
    const videoRef = useRef(null);

    const [testAudioButton, setTestAudioButton] = useState(false);
    const [testVideoButton, setTestVideoButton] = useState(true);

    const [signalingSocket, setSignalingSocket] = useState(null);
    const [peerConnection, setPeerConnection] = useState(null);
    const [dataChannel, setDataChannel] = useState(null);

    const [lastMessage, setLastMessage] = useState(null);

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

            console.log("Server IP:", serverIP);
            console.log("Server Port:", serverPort);

            const peerConnection = new RTCPeerConnection();

            peerConnection.onicecandidate = e => {
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

            peerConnection.oniceconnectionstatechange = function(event) {
                console.log(peerConnection.iceConnectionState);
                if (peerConnection.iceConnectionState === 'connected') {
                    console.log("Closing WebSocket");
                    signalingSocket.close();
                }
            };

            peerConnection.ontrack = function (event) {
                console.log("Got something")
                console.log("new track : " + event.track.kind);
                setVideoStream(event.streams[0]);
            };

            // Create a data channel to send data
            const dataChannel = peerConnection.createDataChannel("data");

            dataChannel.addEventListener("open", () => {
                console.log("Data channel initiated");
            });

            dataChannel.onmessage = function (event) {
                var message = event.data;
                console.log("Received message:");
                console.log(message);
            };

            signalingSocket.onmessage = async (event) => {
                const message = JSON.parse(event.data);
                console.log(message);
                setLastMessage(message);
            };

            setSignalingSocket(signalingSocket);
            setPeerConnection(peerConnection);
            setDataChannel(dataChannel);
            
            console.log("Load Config Success");

        } catch (error) {
            console.error('Error loading configuration:', error);
        }
    }

    // useEffect(() => {
    //     // Log peerConnection whenever it changes
    //     console.log('Updated peerConnection:', peerConnection);
    // }, [peerConnection]); // Run whenever peerConnection changes

    useEffect(() => {
        // Check the last message whenever it changes
        if (lastMessage) {
            console.log('Last message received:', lastMessage);

            if (lastMessage.type === 'offer') {

                console.log("No need for an offer");

            } else if (lastMessage.type === 'answer') {

                handleAnswer(lastMessage.data);

            } else if (lastMessage.type === 'candidate') {

                console.log("No need for a candidate");

            } else {
                console.log('Unknown message type:', lastMessage.type);
            }
        }
    }, [lastMessage]);

    useEffect(() => {
        if (videoRef && videoRef.current && videoRef.current.srcObject) {
            console.log("media stream object created");
            videoRef.current.srcObject = videoStream;
        } else {
            // videoRef.current.srcObject.addTrack(event.track);
            // console.log(videoRef);
            // console.log(videoStream);
            console.log("Can't add stream to video element ...");
        }
    }, [videoStream, videoRef]);

    // useEffect(() => {
    //     console.log(videoRef);
    // }, [videoRef]);
    
    async function handleAnswer(answer) {
        if (!peerConnection) {
            console.error('no peerconnection for answer');
            return;
        }
        console.log("Got answer ?");
        console.log(answer);
        // console.log("Connection was already established with a peer00000000000000000000.");
        if (peerConnection.currentRemoteDescription) {
            console.log("Connection was already established with a peer.");
            return;
        }
        await peerConnection.setRemoteDescription(answer);
    }

    async function initConnection() {
        if (!peerConnection) {
            console.error('no peerconnection for init');
            return;
        }

        if (testAudioButton) {
            peerConnection.addTransceiver('audio', {'direction': 'recvonly'})
        } 
        
        if (testVideoButton) {
            peerConnection.addTransceiver('video', {'direction': 'recvonly'})
        }
    
        const offer = await peerConnection.createOffer();
        signalingSocket.send(JSON.stringify({
            type: 'offer',
            data:  offer
        }));
        await peerConnection.setLocalDescription(offer);
    }

    function send(data) {
        if (!dataChannel) {
            console.error("Data Channel is not initialized.")
            return;
        }
        dataChannel.send(JSON.stringify(data));
    }

    function setupDisplay() {
        console.log("RETRY DISPLAY");
        send({
            type: "DISPLAY",
            data: "",
        });
    }

    // useEffect(() => {
    //     if (peerConnection) {
    //         initConnection();
    //     }
    // }, [peerConnection]);

    // useEffect(() => {
    //     // Call the function to load the configuration
    //     console.log("fhfhfhfhfh")
    //     loadConfig();
    // }, []);

    return { 
        testAudioButton,
        setTestAudioButton,
        testVideoButton,
        setTestVideoButton,
        setupDisplay,
        videoRef,
        setVideoStream,
        videoStream,
        initConnection,
        loadConfig
    };
};

export default StreamHook;