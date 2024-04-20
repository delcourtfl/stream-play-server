// const captureVideo = document.getElementById('captureVideo');

async function startCapture(displayMediaOptions) {
    try {
        return await navigator.mediaDevices
            .getDisplayMedia(displayMediaOptions);
    } catch (err) {
        console.error(err);
        return null;
    }
}


function handleSuccess(stream) {
    const videoElement = document.getElementById('captureVideo');
    videoElement.srcObject = stream;
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