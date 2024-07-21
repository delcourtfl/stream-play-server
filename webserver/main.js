import { 
    connectToSignServer,
    checkVideoReceptionStatus,
    localPeerId,
    hangup,
    makeCall,
    getDataChannel,
    hangupCall,
    startCall,
    joinCall,
    toggleAudio,
    toggleVideo
} from './js/webrtc.js';
import {
    setInputChannel,
    initGamepadInput,
    pauseGamepadInput,
    resumeGamepadInput,
    addGamepad,
    createGamepadMapping

} from './js/inputs.js'; // Adjust the path as needed

await connectToSignServer();

document.getElementById("startButton").addEventListener("click", () => {
    startCall();
});

document.getElementById("callButton").addEventListener("click", () => {
    joinCall();
});

document.getElementById("hangupButton").addEventListener("click", () => {
    hangupCall();
});

document.getElementById("recordVideo").addEventListener("click", () => {
    toggleVideo();
});

document.getElementById("recordAudio").addEventListener("click", () => {
    toggleAudio();
});

document.getElementById("map").addEventListener("click", createGamepadMapping);
document.getElementById("init").addEventListener("click", () => {
    const channel = getDataChannel();
    setInputChannel(channel);
    resumeGamepadInput();
});
document.getElementById("pause").addEventListener("click", pauseGamepadInput);
document.getElementById("addGamepad").addEventListener("click", addGamepad);

setInterval(checkVideoReceptionStatus, 10000);
