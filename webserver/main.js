import { 
    connectToSignServer,
    checkVideoReceptionStatus,
    hangupCall,
    startCall,
    joinCall,
    toggleAudio,
    toggleVideo
} from './js/webrtc.js';
import {
    initGamepadInput,
    toggleGamepadInput,
    handleInputChange,
    createGamepadMapping,
    onGamepadConnected,
    showDefaultMapping,
    onGamepadDisconnected,
    loadSvgObjects
} from './js/inputs.js'; // Adjust the path as needed

await connectToSignServer();

loadSvgObjects();

if ("getGamepads" in navigator) {
    // Start listening for gamepad events
    window.addEventListener("gamepadconnected", onGamepadConnected);
    window.addEventListener("gamepaddisconnected", onGamepadDisconnected);
    showDefaultMapping();
    console.log("Gamepad API is available");
} else {
    console.log("Gamepad API is not supported");
}

initGamepadInput();

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
    const videoElement = document.getElementById('streamVideo');
    // Mute by default to prevent any unintended audio playback
    videoElement.muted = true;
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
document.getElementById("inputrecord").addEventListener("click", toggleGamepadInput);

// Set up an event listener to call the function whenever the input value changes
document.getElementById('controller0Index').addEventListener('input', handleInputChange);
document.getElementById('controller1Index').addEventListener('input', handleInputChange);
document.getElementById('controller2Index').addEventListener('input', handleInputChange);
document.getElementById('controller3Index').addEventListener('input', handleInputChange);

setInterval(checkVideoReceptionStatus, 10000);
