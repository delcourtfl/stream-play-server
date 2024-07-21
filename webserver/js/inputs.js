// array of gamepad elements
var gamepadSvgArr = [];
// buttons elements of the gamepads
var gamepadBtns = [];
var gamepadAnlgs = [];
var SvgLoadedCnt = 0;

var dataChannel = null;

// Flag to control the loop
let isPaused = true;

var currentMappingButtons = [{}, {}, {}, {}]
// format = 0: [true, 1]

var currentMappingAxes = [{}, {}, {}, {}]
// format = 0: [true, 1]

export function setInputChannel(channel) {
    if (!channel) {
        console.log("No proper data channel was provided");
    } else {
        dataChannel = channel;
    }
}

export function send(data) {
    if (dataChannel) {
        dataChannel.send(JSON.stringify(data));
    } else {
        console.log("No datachannel given");
    }
}

async function loadSvgObjects() {
    const svgPromises = [];
    for (let i = 0; i < 4; i++) {
        const obj = document.getElementById("svg-object"+i);
        obj.setAttribute('style', 'filter: contrast(30%);');

        const svgPromise = new Promise((resolve) => {
            obj.addEventListener("load", function() {
                const outerSVGDocument = obj.contentDocument;
                console.log(outerSVGDocument);
    
                var btnArr = [];
                for (let j = 0; j < 16; j++) { 
                    var buttonElement = outerSVGDocument.getElementById("controller-b"+j);
                    btnArr[j] = buttonElement;
                }
                gamepadBtns[i] = btnArr;

                var anlgsArr = [];

                const Lstick = outerSVGDocument.getElementById("controller-b6");
                anlgsArr[0] = Lstick;
                anlgsArr[1] = Lstick;

                const Rstick = outerSVGDocument.getElementById("controller-b7");
                anlgsArr[2] = Rstick;
                anlgsArr[5] = Rstick;

                const Rtrigger = outerSVGDocument.getElementById("controller-a3");
                anlgsArr[3] = Rtrigger;

                const Ltrigger = outerSVGDocument.getElementById("controller-a5");
                anlgsArr[4] = Ltrigger;

                gamepadAnlgs[i] = anlgsArr;
    
                SvgLoadedCnt++;
                resolve(); // Resolve the promise when this SVG object is loaded
            });
    
            // Now that the load event listener is attached, set the data attribute to trigger loading
            obj.setAttribute('data', 'gamepad.svg');
        });
    
        svgPromises.push(svgPromise);

        // Access the contentDocument of the outer <object> (this is the outer SVG document)   
        gamepadSvgArr[i] = obj;
    }

    await Promise.all(svgPromises);
}

loadSvgObjects().then(() => {
    console.log("All SVG objects are loaded and processed.");
    console.log(gamepadBtns);
    
    // Check if the browser supports the Gamepad API
    if ("getGamepads" in navigator) {
        // Start listening for gamepad events
        window.addEventListener("gamepadconnected", onGamepadConnected);
        window.addEventListener("gamepaddisconnected", onGamepadDisconnected);
    } else {
        console.log("Gamepad API is not supported");
    }

    initGamepadInput();
});

// Event handler when a gamepad is connected
function onGamepadConnected(event) {
    const gamepad = event.gamepad;
    const gamepadIndex = gamepad.index;
    const gamepadName = gamepad.id.toString();
    console.log("Gamepad connected:" + gamepad.id + " : " + gamepad.index);

    gamepadSvgArr[gamepad.index].setAttribute('style', '');

    if (currentMappingButtons[gamepadIndex].length > 0) {
        return;
    }

    if (gamepadName.toLowerCase().includes("xbox")){
        console.log("GOT XBOX CONTROLLER");
        // Xbox 360
        currentMappingButtons[gamepadIndex] = {
            0: [false, 12],
            1: [false, 13],
            2: [false, 14],
            3: [false, 15],
            4: [false, 8],
            5: [false, 9],
            6: [true, 3],
            7: [true, 4],
            8: [false, 5],
            9: [false, 4],
            10: [false, 6],
            11: [false, 7],
            12: [false, 0],
            13: [false, 1],
            14: [false, 2],
            15: [false, 3]
        };
        currentMappingAxes[gamepadIndex] = {
            0: [true, 0],
            1: [true, 1],
            2: [true, 2],
            3: [true, 5]
        };
    } else if (gamepadName.toLowerCase().includes("ps3")){
        console.log("GOT PS3 CONTROLLER");
        // PS3
        currentMappingButtons[gamepadIndex] = {
            0: [false, 5],
            1: [false, 6],
            2: [false, 7],
            3: [false, 4],
            4: [false, 0],
            5: [false, 3],
            6: [false, 1],
            7: [false, 2],
            8: [true, 3],
            9: [true, 4],
            10: [false, 8],
            11: [false, 9],
            12: [false, 15],
            13: [false, 13],
            14: [false, 12],
            15: [false, 14]
        };
        currentMappingAxes[gamepadIndex] = {
            0: [true, 0],
            1: [true, 1],
            2: [true, 2],
            5: [true, 5]
        };
    }

    console.log(currentMappingButtons[gamepadIndex]);
    console.log(currentMappingAxes[gamepadIndex]);

    const table = document.getElementById("defaultcontroller");

    while (table.rows.length > 0) {
        table.deleteRow(0);
    }

    for (const element of defaultXboxController) {
        const row = table.insertRow();
        const cell1 = row.insertCell(0);
        const cell2 = row.insertCell(1);
        const cell3 = row.insertCell(2);

        cell1.innerHTML = element.name;
        cell2.innerHTML = element.value;
        cell3.innerHTML = element.isAnalog;
    }

    populateTable(gamepadIndex);
    
}

// Event handler when a gamepad is disconnected
function onGamepadDisconnected(event) {
    const gamepad = event.gamepad;
    console.log("Gamepad disconnected:" + gamepad.id + " : " + gamepad.index);
    gamepadSvgArr[gamepad.index].setAttribute('style', 'filter: contrast(30%);');
}

export function initGamepadInput() {
    console.log("INIT GAMEPAD INPUT");

    // Set the bit at the specified index (buttonIndex)
    function setBit(buttonIndex) {
        nextBitArray |= 1 << buttonIndex;
    }

    // Initialize the bit array
    let bitArray = [0, 0, 0, 0];
    let nextBitArray = 0;
    let stickValues = [[0, 0, 0, 0, 0, 0], [0, 0, 0, 0, 0, 0], [0, 0, 0, 0, 0, 0], [0, 0, 0, 0, 0, 0]];
    let moved = [false, false, false, false];

    // Function to update gamepad input
    function updateGamepadInput() {
        if (isPaused) {
            // Exit the function if it's paused
            // console.log("Stop input recording")
            return;
        }
        // Get the list of connected gamepads
        const gamepads = navigator.getGamepads();

        // Iterate through each gamepad
        for (let i = 0; i < gamepads.length; i++) {
            const gamepad = gamepads[i];
            if (gamepad) {
                nextBitArray = 0;
                moved[i] = false;                        

                // Handle gamepad buttons input
                for (let j = 0; j < gamepad.buttons.length; j++) {
                    if (j in currentMappingButtons[i]){
                        const button = gamepad.buttons[j];
                        const curr = currentMappingButtons[i][j];
                        // Check if the button is pressed
                        if (button.pressed) {
                            if (!curr[0]) {
                                setBit(curr[1], nextBitArray);
                            } else {
                                const axisValue = floorToNearest0_1(button.value);
                                if (stickValues[i][curr[1]] !== axisValue) {
                                    stickValues[i][curr[1]] = axisValue;
                                    moved[i] = true;
                                    displayAnalogInput(i, curr[1], axisValue);
                                }
                            }
                        } else {
                            if (curr[0] && stickValues[i][curr[1]] !== 0){
                                stickValues[i][curr[1]] = 0;
                                moved[i] = true;
                                displayAnalogInput(i, curr[1], 0);
                            }
                        }
                    }
                }

                // Handle gamepad axis input
                for (let k = 0; k < gamepad.axes.length; k++) {
                    if (k in currentMappingAxes[i]){
                        const axis = floorToNearest0_1(gamepad.axes[k]);
                        const curr = currentMappingAxes[i][k];
                        // Check if the axe is moved
                        if (!curr[0] && Math.abs(axis) > 0.2) {
                            setBit(curr[1], nextBitArray);
                        } else if (axis !== stickValues[i][curr[1]]) {
                            stickValues[i][curr[1]] = axis;
                            moved[i] = true;
                            displayAnalogInput(i, curr[1], axis);
                        }
                    }
                }

                if (nextBitArray !== bitArray[i]) {
                    moved[i] = true;
                    bitArray[i] = nextBitArray;

                    for (let j = 0; j < 16; j++) { 
                        var buttonElement = gamepadBtns[i][j];
                        if (buttonElement) {
                            if (bitArray[i] & (1 << j)) {
                                buttonElement.setAttribute('style', 'fill: blue;');
                            } else {
                                buttonElement.setAttribute('style', '');
                            }
                        }
                    }
                }

                if (moved[i]) {
                    console.log("Input:", bitArray[i].toString(2), stickValues);
                    // if (signalingSocket && signalingSocket.readyState === WebSocket.OPEN) {
                    send({
                        type: "GAMEPAD",
                        data: JSON.stringify({
                            wButtons: bitArray[i],
                            bLeftTrigger: stickValues[i][3] * 255 | 0,
                            bRightTrigger: stickValues[i][4] * 255 | 0,
                            sThumbLX: stickValues[i][0] * 32767 | 0,
                            sThumbLY: stickValues[i][1] * -32767 | 0,
                            sThumbRX: stickValues[i][2] * 32767 | 0,
                            sThumbRY: stickValues[i][5] * -32767 | 0,
                        }),
                        option: i,
                    });
                    // } else {
                    //     console.log("WebSocket is not open. Data not sent.");
                    // }
                }
            }
        }

        // Continue listening for gamepad input
        requestAnimationFrame(updateGamepadInput);
    }

    // Start listening for gamepad input
    requestAnimationFrame(updateGamepadInput);
}

function displayAnalogInput(gamepadIndex, index, value) {
    const multiplier = 25;
    switch (index) {
        case 0:
            const Lx = Number(gamepadAnlgs[gamepadIndex][0].dataset.originalXPosition);
            gamepadAnlgs[gamepadIndex][0].setAttribute("cx", Lx + value * multiplier);
            break;
        case 1:
            const Ly = Number(gamepadAnlgs[gamepadIndex][1].dataset.originalYPosition);
            gamepadAnlgs[gamepadIndex][1].setAttribute("cy", Ly + value * multiplier);
            break;
        case 2:
            const Rx = Number(gamepadAnlgs[gamepadIndex][2].dataset.originalXPosition);
            gamepadAnlgs[gamepadIndex][2].setAttribute("cx", Rx + value * multiplier);
            break;
        case 3:
            if (value > 0.2) {
                gamepadAnlgs[gamepadIndex][3].setAttribute('style', 'fill: blue;');
            } else {
                gamepadAnlgs[gamepadIndex][3].setAttribute('style', '');
            }
            break;
        case 4:
            if (value > 0.2) {
                gamepadAnlgs[gamepadIndex][4].setAttribute('style', 'fill: blue;');
            } else {
                gamepadAnlgs[gamepadIndex][4].setAttribute('style', '');
            }
            break;
        case 5:
            const Ry = Number(gamepadAnlgs[gamepadIndex][5].dataset.originalYPosition);
            gamepadAnlgs[gamepadIndex][5].setAttribute("cy", Ry + value * multiplier);
            break;
        default:
            console.log("Shoud not happen, bad index for analog buttons");
    }
}

function floorToNearest0_1(value) {
    return Math.floor(value * 10) / 10;
}

// Function to pause the gamepad input
export function pauseGamepadInput() {
    isPaused = true;
}

// Function to resume the gamepad input
export function resumeGamepadInput() {
    // isPaused = false;
    if (isPaused) {
        isPaused = false;
        initGamepadInput(); // Start/resume the loop
    }
}

const defaultXboxController = [
    { name: "DPadUp", value: 0, isAnalog: false },
    { name: "DPadDown", value: 1, isAnalog: false },
    { name: "DPadLeft", value: 2, isAnalog: false },
    { name: "DPadRight", value: 3, isAnalog: false },
    { name: "Start", value: 4, isAnalog: false },
    { name: "Back", value: 5, isAnalog: false },
    { name: "LeftThumbstick", value: 6, isAnalog: false },
    { name: "RightThumbstick", value: 7, isAnalog: false },
    { name: "LeftShoulder", value: 8, isAnalog: false },
    { name: "RightShoulder", value: 9, isAnalog: false },
    // { name: "Guide", value: 10, isAnalog: false },
    { name: "A", value: 12, isAnalog: false },
    { name: "B", value: 13, isAnalog: false },
    { name: "X", value: 14, isAnalog: false },
    { name: "Y", value: 15, isAnalog: false },
    { name: "LeftTrigger", value: 3, isAnalog: true },
    { name: "RightTrigger", value: 4, isAnalog: true },
    { name: "LeftThumbstick (X-Axis)", value: 0, isAnalog: true },
    { name: "LeftThumbstick (Y-Axis)", value: 1, isAnalog: true },
    { name: "RightThumbstick (X-Axis)", value: 2, isAnalog: true },
    { name: "RightThumbstick (Y-Axis)", value: 5, isAnalog: true }
];

export async function createGamepadMapping() {

    // Get the list of connected gamepads
    const gamepads = navigator.getGamepads();

    // Prompt the user to press each button
    console.log("Press each gamepad button when prompted:");

    // Iterate through each gamepad
    for (let i = 0; i < gamepads.length; i++) {
        const gamepad = gamepads[i];
        if (gamepad) {
            console.log(i);
            await checkButtonPress(gamepad); // Start checking for button press
        }
    }

    // Function to handle button mapping
    async function mapButtons(gamepadIndex, askedInput, defaultValues, isAnalog) {
        return new Promise((resolve, reject) => {
        const gamepad = navigator.getGamepads()[gamepadIndex];

        const buttons = gamepad.buttons;
        for (let i = 0; i < buttons.length; i++) {
            const button = buttons[i];

            if (button.pressed) {
                console.log(i + ", " + isAnalog + ", " + askedInput);
                currentMappingButtons[gamepadIndex][i] = [isAnalog, askedInput];
                resolve(); // Resolve the Promise to signal completion
                return; // Exit the loop and function
            }
        }

        const axes = gamepad.axes;
        for (let i = 0; i < axes.length; i++) {
            const axe = axes[i];

            if (Math.abs(axe - defaultValues[i]) > 0.5) {
                // console.log(axe);
                console.log(i + ", " + isAnalog + ", " + askedInput);
                currentMappingAxes[gamepadIndex][i] = [isAnalog, askedInput];
                resolve(); // Resolve the Promise to signal completion
                return; // Exit the loop and function
            }
        }

        // If no button is pressed, schedule the next iteration
        requestAnimationFrame(() => {
            mapButtons(gamepadIndex, askedInput, defaultValues, isAnalog)
            .then(resolve) // Propagate the resolve signal
            .catch(reject); // Propagate any error
            });
        });
    }

    // Function to continuously check for button press
    async function checkButtonPress(gamepad) {
        // const gamepad = navigator.getGamepads()[gamepadindex];
        const axes = gamepad.axes;
        const gamepadindex = gamepad.index;
        var defaultValues = [];

        for (let i = 0; i < axes.length; i++) {
            defaultValues[i] = axes[i];
        }

        for (const element of defaultXboxController) {
            console.log(`Key: ${element.name}, Value: ${element.value}, IsAnalog: ${element.isAnalog}` );
            try {
                await mapButtons(gamepadindex, element.value, defaultValues, element.isAnalog);
                console.log("Is ok");
                await delay(1000);
            } catch (error) {
                console.log("Error mapping")
            }
        }

        console.log(currentMappingButtons[gamepadindex]);
        console.log(currentMappingAxes[gamepadindex]);
        console.log(gamepadindex);
        populateTable(gamepadindex);
    }
}

// Function to introduce a delay using setTimeout
function delay(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
}

// Function to populate the table with data
function populateTable(index) {
    const table = document.getElementById("controller"+index);

    while (table.rows.length > 0) {
        table.deleteRow(0);
    }

    // For currentMappingButtons
    for (const [key, value] of Object.entries(currentMappingButtons[index])) {
        const row = table.insertRow();
        const cell1 = row.insertCell(0);
        cell1.innerHTML = key;
        const cell2 = row.insertCell(1);
        cell2.innerHTML = value;
    }

    // For currentMappingAxes
    for (const [key, value] of Object.entries(currentMappingAxes[index])) {
        const row = table.insertRow();
        const cell1 = row.insertCell(0);
        cell1.innerHTML = key;
        const cell2 = row.insertCell(1);
        cell2.innerHTML = value;
    }

}

export function addGamepad() {
    console.log("Adding gamepad on the server");
    send({
        type: "ADDGAMEPAD",
        data: "",
    });
}
