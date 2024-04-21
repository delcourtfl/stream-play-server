import React, { useEffect } from 'react';

const PlayTab = ({ streamHook }) => {

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

export default PlayTab;