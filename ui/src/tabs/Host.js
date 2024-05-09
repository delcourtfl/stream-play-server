import React, { useEffect } from 'react';
import Button from '@mui/joy/Button';
import Box from '@mui/joy/Box';

const HostMedia = ({ castHook }) => {
    
    async function startCapture(displayMediaOptions) {
        try {
            return await navigator.mediaDevices.getDisplayMedia(displayMediaOptions);
        } catch (err) {
            console.error(err);
            return null;
        }
    }

    const setLocalStream = () => {
        if (castHook.captureRef && castHook.captureRef.current && castHook.captureStream) {
            castHook.captureRef.current.srcObject = castHook.captureStream;
            console.log("Local video is set");
        }
    }

    const setRemoteStream = () => {
        if (castHook.videoRef && castHook.videoRef.current && castHook.videoStream) {
            castHook.videoRef.current.srcObject = castHook.videoStream;
            console.log("Remote video is set");
        }
    }

    async function startCall() {
        const displayMediaOptions = { video: true }; // You can customize options here

        const stream = await startCapture(displayMediaOptions);
        castHook.setCaptureStream(stream);
        if (stream) {
            setLocalStream();
            castHook.setStartDisabled(true);
            castHook.setHangupDisabled(false);
            castHook.signalingSocket.send(JSON.stringify({
                type: 'ready'
            }));
        } else {
            console.error("Failed to capture media stream.");
        }
    }

    useEffect(() => {
        if (castHook.captureStream) {
            setLocalStream();
        }
        if (castHook.videoStream) {
            setRemoteStream();
        }
    }, []);

    useEffect(() => {
        if (castHook.captureStream) {
            setLocalStream();
        }
        if (castHook.videoStream) {
            setRemoteStream();
        }
    }, [castHook.captureStream, castHook.videoStream]);

    return (
        <Box
            sx={{
                width: '100vw',
                height: '100vh',
                position: 'fixed',
                top: 0,
                left: 0,
                objectFit: 'contain', // Ensure video takes full space
            }}
        >
            <Button id="loadButton" disabled={!!castHook.signalingSocket} onClick={castHook.loadConfig}>Load Config</Button>
            <Button id="startButton" disabled={castHook.startDisabled} onClick={startCall}>Start</Button>
            <Button id="hangupButton" disabled={castHook.hangupDisabled} onClick={castHook.hangupCall}>Hang Up</Button>
            {/* <Button fullWidth={true} size='lg' onClick={startCaptureAndDisplay} sx={{height: '5%'}}>Start Capture</Button> */}
            <video ref={castHook.captureRef} id="localVideo" autoPlay controls style={{ width: '100%', height: '45%', objectFit: 'contain' }} />
            <video ref={castHook.videoRef} id="remoteVideo" autoPlay controls style={{ width: '100%', height: '45%', objectFit: 'contain' }} />
            {/* <div> */}
            {/* <video id="localVideo" autoPlay muted srcObject={localStream}></video> */}
            {/* <video id="remoteVideo" autoPlay></video> */}
        </Box>
    );
};

export default HostMedia;