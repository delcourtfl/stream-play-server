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
    createGamepadMapping,
    onGamepadConnected,
    onGamepadDisconnected,
    loadSvgObjects
} from './js/inputs.js'; // Adjust the path as needed

await connectToSignServer();

await loadSvgObjects().then(() => {    
    // Check if the browser supports the Gamepad API
    if ("getGamepads" in navigator) {
        // Start listening for gamepad events
        window.addEventListener("gamepadconnected", onGamepadConnected);
        window.addEventListener("gamepaddisconnected", onGamepadDisconnected);
        console.log("Gamepad API is available");
    } else {
        console.log("Gamepad API is not supported");
    }

    initGamepadInput();
});

// Get the button element
const hostButton = document.getElementById('hostButton');
// Get the current hostname
const currentHostname = window.location.hostname;
console.log(currentHostname);

// Check if the hostname is 'localhost'
if (currentHostname !== 'localhost') {
    // Disable the button if the hostname is not 'localhost'
    hostButton.disabled = true;
    hostButton.textContent = "Host (localhost only)";
} else {
    // Enable the button if the hostname is 'localhost'
    hostButton.disabled = false;
}

document.getElementById("hostButton").addEventListener("click", () => {
    startCall();
});

document.getElementById("joinButton").addEventListener("click", () => {
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
