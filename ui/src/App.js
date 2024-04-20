import React, { useEffect } from 'react';
import { Button, Tab, Tabs } from '@mui/material';

function App() {
    useEffect(() => {
        // Function to handle opening tabs
        const openTab = (event, tabName) => {
            const tabContents = document.getElementsByClassName('tabcontent');
            for (let i = 0; i < tabContents.length; i++) {
                tabContents[i].style.display = 'none';
            }
            const tabLinks = document.getElementsByClassName('tablinks');
            for (let i = 0; i < tabLinks.length; i++) {
                tabLinks[i].className = tabLinks[i].className.replace(' active', '');
            }
            document.getElementById(tabName).style.display = 'block';
            event.currentTarget.className += ' active';
        };
        document.getElementById('default').click();

        // Override console.log and console.error
        const originalConsoleLog = console.log;
        const originalConsoleError = console.error;
        console.log = (...args) => {
            const message = args.join(' ');
            const chatMessage = document.createElement('div');
            chatMessage.classList.add('chat-message');
            chatMessage.textContent = 'Browser: ' + message;
            document.getElementById('logContainer').appendChild(chatMessage);
            originalConsoleLog(...args);
        };
        console.error = (...args) => {
            const message = args.join(' ');
            const chatMessage = document.createElement('div');
            chatMessage.classList.add('chat-message');
            chatMessage.textContent = message;
            document.getElementById('logContainer').appendChild(chatMessage);
            originalConsoleError(...args);
        };
    }, []);

    return (
        <div>
            <Tabs>
                <Tab label="Game" onClick={(e) => openTab(e, 'Game')} id="default"></Tab>
                <Tab label="Controllers" onClick={(e) => openTab(e, 'Controllers')}></Tab>
                <Tab label="Connection" onClick={(e) => openTab(e, 'Connection')}></Tab>
                <Tab label="Capture" onClick={(e) => openTab(e, 'Capture')}></Tab>
            </Tabs>

            <div id="Game" className="tabcontent">
                <video id="video" controls autoPlay muted playsInline></video>
            </div>

            {/* Other tab contents */}

            <div id="logContainer" className="chat-window">
            <div className="chat-message">Browser: Hello there!</div>
            </div>
        </div>
    );
}

export default App;
