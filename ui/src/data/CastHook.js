import { useState, useEffect, useRef } from 'react';

const CastHook = () => {

    const [captureStream, setCaptureStream] = useState(null);
    const captureRef = useRef(null);

    const [testAudioButton, setTestAudioButton] = useState(false);
    const [testVideoButton, setTestVideoButton] = useState(true);

    const [signalingSocket, setSignalingSocket] = useState(null);
    const [peerConnection, setPeerConnection] = useState(null);
    const [dataChannel, setDataChannel] = useState(null);

    const [lastMessage, setLastMessage] = useState(null);

    // const setStream = (stream) => {
    //     setVideoStream(stream);
    // };

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

            // captureStream.getTracks().forEach(track => peerConnection.addTrack(track, captureStream));
            // console.log(captureStream.getTracks());
            // console.log(peerConnection)

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

    useEffect(() => {
        // Check the last message whenever it changes
        if (lastMessage) {
            console.log('Last message received:', lastMessage);

            if (lastMessage.type === 'offer') {

                console.log('Received offer');
                handleOffer(lastMessage.data);

            } else if (lastMessage.type === 'answer') {

                // handleAnswer(lastMessage.data);
                console.log("No need for answer");

            } else if (lastMessage.type === 'candidate') {

                console.log('Received ICE candidate');
                handleCandidate(lastMessage.data);

            } else {
                console.log('Unknown message type:', lastMessage.type);
            }
        }
    }, [lastMessage]);

    // async function createPeerConnection() {
    //     const peerConnection = new RTCPeerConnection();
    //     // Create a data channel to send data
    //     const dataChannel = peerConnection.createDataChannel("data");

    //     dataChannel.addEventListener("open", () => {
    //         console.log("Data channel initiated");
    //     });

    //     dataChannel.onmessage = function(event) {
    //         var message = event.data;
    //         console.log("Received message:");
    //         console.log(message);
    //     };

    //     peerConnection.onicecandidate = e => {
    //         const message = {
    //             type: 'candidate',
    //             candidate: null,
    //         };
    //         if (e.candidate) {
    //             message.candidate = e.candidate.candidate;
    //             message.sdpMid = e.candidate.sdpMid;
    //             message.sdpMLineIndex = e.candidate.sdpMLineIndex;
    //         }
    //         // signalingSocket.postMessage(message);
    //         signalingSocket.send(JSON.stringify(message));
    //     };
    //     if (videoRef.current) {
    //         peerConnection.ontrack = e => videoRef.current.srcObject = e.streams[0];
    //     }
    //     if (captureStream) {
    //         captureStream.getTracks().forEach(track => peerConnection.addTrack(track, captureStream));
    //     }

    //     // setPeerConnection(peerConnection);
    //     // setDataChannel(dataChannel);
    //     // return peerConnection;
    // }

    async function handleOffer(offer) {
        if (!peerConnection) {
            console.error('no peerconnection for offer');
            return;
        }

        captureStream.getTracks().forEach(track => peerConnection.addTrack(track, captureStream));

        await peerConnection.setRemoteDescription(offer);
        
        const answer = await peerConnection.createAnswer();
        signalingSocket.send(JSON.stringify({
            type: 'answer',
            data:  answer
        }));
        await peerConnection.setLocalDescription(answer);
    }
    
    // async function handleAnswer(answer) {
    //     if (!peerConnection) {
    //         console.error('no peerconnection for answer');
    //         return;
    //     }
    //     if (peerConnection.currentRemoteDescription) {
    //         console.log("Connection was already established with a peer.")
    //     }
    //     await peerConnection.setRemoteDescription(answer);
    // }
    
    async function handleCandidate(candidate) {
        if (!peerConnection) {
            console.error('no peerconnection for candidate');
            return;
        }
        if (!candidate || !candidate.candidate) {
            await peerConnection.addIceCandidate(null);
        } else {
            await peerConnection.addIceCandidate(candidate);
        }
    }

    // async function initConnection() {
    //     if (!peerConnection) {
    //         console.error('no peerconnection for init');
    //         return;
    //     }

    //     // if (testAudioButton) {
    //     //     peerConnection.addTransceiver('audio', {'direction': 'recvonly'})
    //     // } 
        
    //     // if (testVideoButton) {
    //     //     peerConnection.addTransceiver('video', {'direction': 'recvonly'})
    //     // }
    
    //     const offer = await peerConnection.createOffer();
    //     signalingSocket.send(JSON.stringify({
    //         type: 'offer',
    //         data:  offer
    //     }));
    //     await peerConnection.setLocalDescription(offer);
    // }

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
    //     initConnection();
    // }, [captureStream]);

    useEffect(() => {
        // Call the function to load the configuration
        if (captureStream) {
            loadConfig();
            // initConnection();
        }
    }, [captureStream]);

    return { 
        // setStream,
        testAudioButton,
        setTestAudioButton,
        testVideoButton,
        setTestVideoButton,
        setupDisplay,
        // setupPeerConnection,
        //
        captureRef,
        setCaptureStream,
        captureStream,
        //
        // videoRef,
        // setVideoStream,
        // videoStream,
        //
        // initConnection,
    };
};

export default CastHook;