import { useState, useEffect, useRef } from 'react';

const StreamHook = () => {
    const [videoStream, setVideoStream] = useState(null);
    const videoRef = useRef(null);

    // const [captureStream, setCaptureStream] = useState(null);
    // const captureRef = useRef(null);

    const [testAudioButton, setTestAudioButton] = useState(false);
    const [testVideoButton, setTestVideoButton] = useState(true);

    const [signalingSocket, setSignalingSocket] = useState(null);
    const [peerConnection, setPeerConnection] = useState(null);
    const [dataChannel, setDataChannel] = useState(null);

    const [lastMessage, setLastMessage] = useState(null);

    const setStream = (stream) => {
        setVideoStream(stream);
    };

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

                // if (videoRef && videoRef.current && videoRef.current.srcObject) {
                //     console.log("media stream object created");
                //     videoRef.current.srcObject = event.streams[0];
                // } else {
                //     // videoRef.current.srcObject.addTrack(event.track);
                //     console.log(videoRef);
                //     console.log("Can't add tracks ...");
                // }
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

    useEffect(() => {
        // Log peerConnection whenever it changes
        console.log('Updated peerConnection:', peerConnection);
    }, [peerConnection]); // Run whenever peerConnection changes

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
            console.log(videoRef);
            console.log(videoStream);
            console.log("Can't add tracks ...");
        }
    }, [videoStream, videoRef]);

    useEffect(() => {
        console.log(videoRef);
    }, [videoRef]);

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

    // async function handleOffer(offer) {
    //     if (!peerConnection) {
    //         console.error('no peerconnection for offer');
    //         return;
    //     }
    //     await peerConnection.setRemoteDescription(offer);
        
    //     const answer = await peerConnection.createAnswer();
    //     signalingSocket.send(JSON.stringify({
    //         type: 'answer',
    //         data:  answer
    //     }));
    //     await peerConnection.setLocalDescription(answer);
    // }
    
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
    
    // async function handleCandidate(candidate) {
    //     if (!peerConnection) {
    //         console.error('no peerconnection for candidate');
    //         return;
    //     }
    //     if (!candidate.candidate) {
    //         await peerConnection.addIceCandidate(null);
    //     } else {
    //         await peerConnection.addIceCandidate(candidate);
    //     }
    // }

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

    // function setupPeerConnection() {
    //     console.log("Sending Offer");

    //     peerConnection.oniceconnectionstatechange = function(event) {
    //         console.log(peerConnection.iceConnectionState);
    //         if (peerConnection.iceConnectionState === 'connected') {
    //             console.log("Closing WebSocket");
    //             signalingSocket.close();
    //             // initInput();
    //         }
    //     };

    //     peerConnection.onicecandidate = handleIceCandidate;

    //     peerConnection.ontrack = function (event) {
    //         console.log("new track : " + event.track.kind);

    //         if (!videoRef.current.srcObject) {
    //             console.log("media stream object created");
    //             videoRef.current.srcObject = event.streams[0];
    //         } else {
    //             videoRef.current.srcObject.addTrack(event.track);
    //         }
    //     };

    //     if (testAudioButton) {
    //         peerConnection.addTransceiver('audio', {'direction': 'recvonly'})
    //     } 
        
    //     if (testVideoButton) {
    //         peerConnection.addTransceiver('video', {'direction': 'recvonly'})
    //     }

    //     peerConnection.createOffer()
    //         .then(offer => {
    //             peerConnection.setLocalDescription(offer);
    //             signalingSocket.send(JSON.stringify({
    //                 type: 'offer',
    //                 data:  offer
    //             }));
    //         });

    //     // Function to check video reception status
    //     function checkVideoReceptionStatus() {
    //         peerConnection.getStats().then((stats) => {
    //             printRTCStatsReport(stats);
    //             stats.forEach((report) => {
    //                 if (report.type === 'inbound-rtp' && report.kind === 'video') {
    //                     console.log('Video data received: ' + report.framesReceived + ' frames');
    //                 }
    //             });
    //         });
    //     }

    //     // Function to print RTCStatsReport cleanly using console.table()
    //     function printRTCStatsReport(report) {
    //         const reportEntries = [];
            
    //         report.forEach((value, key) => {
    //             // Skip entries that are not objects (e.g., report ID)
    //             if (typeof value !== 'object' || value === null) {
    //                 return;
    //             }
            
    //             const entry = {
    //                 Type: value.type,
    //                 ID: key,
    //             };
            
    //             // Extract and add properties to the entry
    //             for (const prop in value) {
    //                 if (prop !== 'type') {
    //                 entry[prop] = value[prop];
    //                 }
    //             }
        
    //             reportEntries.push(entry);
    //         });
            
    //         // Print the report using console.table()
    //         // console.table(reportEntries);
    //         console.log(reportEntries);
    //     }
            
    //     // Call the function periodically to monitor video reception
    //     setInterval(checkVideoReceptionStatus, 10000); // Check every second (adjust as needed)
    // }

    function setupDisplay() {
        console.log("RETRY DISPLAY");
        send({
            type: "DISPLAY",
            data: "",
        });
    }

    useEffect(() => {
        // Call the function to load the configuration
        loadConfig();
    }, []);

    return { 
        // setStream,
        testAudioButton,
        setTestAudioButton,
        testVideoButton,
        setTestVideoButton,
        setupDisplay,
        // setupPeerConnection,
        //
        // captureRef,
        // setCaptureStream,
        // captureStream,
        //
        videoRef,
        setVideoStream,
        videoStream,
        //
        initConnection,
    };
};

export default StreamHook;