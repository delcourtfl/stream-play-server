import React from 'react';

const ClientPlayer = ({ streamHook }) => {
    return (
        <video
            // src={src}
            autoPlay
            controls
            style={{
                width: '100%',
                height: '100%',
                objectFit: 'cover', // Ensures video fills the container
            }}
        />
    );
};

export default ClientPlayer;