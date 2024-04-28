import {useState, useRef} from 'react';
import Button from '@mui/joy/Button';

const ClientConnectionTab = ({ streamHook }) => {
    // const [testAudioButton, setTestAudioButton] = useState(false);
    // const [testVideoButton, setTestVideoButton] = useState(true);
    const logContainerRef = useRef(null);

    const testAudio = () => {
        streamHook.setTestAudioButton(!streamHook.testAudioButton);
    };

    const testVideo = () => {
        streamHook.setTestVideoButton(!streamHook.testVideoButton);
    };

    return (
        <div id="Connection" className="tabcontent">
            <Button id="connectPC" onClick={streamHook.initConnection}> Connect </Button>
            <Button id="retryFFMPEG" onClick={streamHook.setupDisplay}> DisplayRetry </Button>
            <Button id="testAudio" onClick={testAudio}> Audio : {streamHook.testAudioButton ? 'true' : 'false'}</Button>
            <Button id="testVideo" onClick={testVideo}> Video : {streamHook.testVideoButton ? 'true' : 'false'}</Button> 
            <div id="logContainer" className="chat-window" ref={logContainerRef}>
                <div className="chat-message">Browser : Hello there!</div>
                {/* More chat messages */}
            </div>
        </div>
    );
};

export default ClientConnectionTab;