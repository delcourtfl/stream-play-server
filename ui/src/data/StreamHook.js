import { useState } from 'react';

const StreamHook = () => {
    const [videoStream, setVideoStream] = useState(null);

    const setStream = (stream) => {
        setVideoStream(stream);
    };

    return { videoStream, setStream };
};

export default StreamHook;