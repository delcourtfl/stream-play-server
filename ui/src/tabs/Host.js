import React, { useEffect } from 'react';
import Button from '@mui/joy/Button';

const HostMedia = ({ streamHook }) => {
    
    async function startCapture(displayMediaOptions) {
        try {
            return await navigator.mediaDevices.getDisplayMedia(displayMediaOptions);
        } catch (err) {
            console.error(err);
            return null;
        }
    }

    function handleSuccess(stream) {
        const videoElement = document.getElementById('captureVideo');
        if (videoElement) {
            videoElement.srcObject = stream;
        }
        console.log("end");
    }

    async function startCaptureAndDisplay() {
        const displayMediaOptions = { video: true }; // You can customize options here

        const stream = await startCapture(displayMediaOptions);
        if (stream) {
            handleSuccess(stream);
        } else {
            console.error("Failed to capture media stream.");
        }
    }

    return (
        <div>
            <video id="captureVideo" autoPlay controls style={{ width: '100%', height: 'auto' }} />
            <Button fullWidth={true} size='lg' onClick={startCaptureAndDisplay}>Start Capture</Button>
        </div>
    );
};

export default HostMedia;