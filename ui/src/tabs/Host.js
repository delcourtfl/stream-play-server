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

    const setStreamAsSource = (stream) => {
        if (castHook.captureRef && castHook.captureRef.current && stream) {
            castHook.captureRef.current.srcObject = stream;
            // Start sending on the webrtc connection
            console.log("Cast Mirror is set");
        }
    }

    async function startCaptureAndDisplay() {
        const displayMediaOptions = { video: true }; // You can customize options here

        const stream = await startCapture(displayMediaOptions);
        castHook.setCaptureStream(stream);
        if (stream) {
            setStreamAsSource(stream);
        } else {
            console.error("Failed to capture media stream.");
        }
    }

    useEffect(() => {
        if (castHook.captureStream) {
            setStreamAsSource(castHook.captureStream)
        }
    }, []);

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
            <Button fullWidth={true} size='lg' onClick={startCaptureAndDisplay} sx={{height: '5%'}}>Start Capture</Button>
            <video ref={castHook.captureRef} id="captureVideo" autoPlay controls style={{ width: '100%', height: '45%', objectFit: 'contain' }} />
            <video ref={castHook.captureRef2} id="captureVideo2" autoPlay controls style={{ width: '100%', height: '45%', objectFit: 'contain' }} />
        </Box>
    );
};

export default HostMedia;