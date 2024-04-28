import React, { useEffect } from 'react';

const PlayTab = ({ streamHook }) => {

    const setStreamAsSource = (stream) => {
        if (streamHook.videoRef && streamHook.videoRef.current && stream) {
            streamHook.videoRef.current.srcObject = stream;
            // Start sending on the webrtc connection
            console.log("Stream is set");
        }
    }

    useEffect(() => {
        if (streamHook.videoStream) {
            setStreamAsSource(streamHook.videoStream)
        }
    }, []);

    return (
        <video
            ref={streamHook.videoRef}
            autoPlay
            controls
            style={{
                width: '100vw',
                height: '100vh',
                position: 'fixed',
                top: 0,
                left: 0,
                objectFit: 'contain', // Ensure video takes full space
            }}
        />
    );
};

export default PlayTab;