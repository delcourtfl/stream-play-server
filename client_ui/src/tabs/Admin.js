import React, { useEffect, useState } from 'react';
import { Typography, Paper } from '@mui/joy';

import Table from '@mui/joy/Table';

const AdminPage = ({ streamHook }) => {
    const [logs, setLogs] = useState({
        video: '',
        audio: '',
        client: '',
        server: '',
        sign: ''
    });

    const fetchLogData = async () => {
        try {
            const logTypes = ['video', 'audio', 'client', 'server', 'sign'];
            const updatedLogs = {};

            for (const logType of logTypes) {
                const response = await fetch(`/admin/logs/${logType}`);
                const data = await response.text();
                updatedLogs[logType] = data;
            }

            setLogs(updatedLogs);
        } catch (error) {
            console.error('Error fetching log data:', error);
        }
    };

    useEffect(() => {

        fetchLogData();

        // Fetch log data every 10 seconds
        const interval = setInterval(fetchLogData, 10000);

        return () => clearInterval(interval);

    }, []);

    return (
        <div>
            <Typography variant="h1" gutterBottom>
                Stream Play Server : Admin Logs
            </Typography>

            <Table>
                <thead>
                    <tr>
                    <th>Log 1</th>
                    </tr>
                </thead>
                <tbody>
                    <tr key={"VideoLogs"}>
                        <td>{logs['video']}</td>
                    </tr>
                </tbody>
            </Table>

            <Table>
                <thead>
                    <tr>
                    <th>Log 2</th>
                    </tr>
                </thead>
                <tbody>

                </tbody>
            </Table>

            <Table>
                <thead>
                    <tr>
                    <th>Log 3</th>
                    </tr>
                </thead>
                <tbody>
                    
                </tbody>
            </Table>
        </div>
    );
}

export default AdminPage;
