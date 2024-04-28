import * as React from 'react';
import Box from '@mui/joy/Box';
import ListItemDecorator from '@mui/joy/ListItemDecorator';
import Tabs from '@mui/joy/Tabs';
import TabList from '@mui/joy/TabList';
import Tab, { tabClasses } from '@mui/joy/Tab';
import ComputerIcon from '@mui/icons-material/Computer';
import PlayArrowIcon from '@mui/icons-material/PlayArrow';
import ConnectWithoutContactIcon from '@mui/icons-material/ConnectWithoutContact';
import SportsEsportsIcon from '@mui/icons-material/SportsEsports';
import SettingsIcon from '@mui/icons-material/Settings';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import ExpandLessIcon from '@mui/icons-material/ExpandLess';
import TabPanel from '@mui/joy/TabPanel';
import Button from '@mui/joy/Button';

import AdminPage from './tabs/Admin';
import ClientPlayer from'./tabs/Client';
import HostMedia from './tabs/Host';
import ControllersTab from './tabs/Controllers'
import PlayTab from './tabs/Game';

import StreamHook from './data/StreamHook';
import CastHook from './data/CastHook';

export default function App() {
    const [index, setIndex] = React.useState(0);
    const colors = ['primary', 'primary', 'primary', 'warning', 'warning'];
    const [collapsed, setCollapsed] = React.useState(false);

    const handleCollapseToggle = () => {
        setCollapsed(!collapsed);
    };

    const streamHook = StreamHook();
    const castHook = CastHook();

    return (
        <Tabs
            size="lg"
            aria-label="Bottom Navigation"
            value={index}
            onChange={(event, value) => setIndex(value)}
            sx={(theme) => ({
                // p: 1,
                borderRadius: 16,
                // maxWidth: 400,
                // mx: 'auto',
                boxShadow: theme.shadow.sm,
                // '--joy-shadowChannel': theme.vars.palette[colors[index]].darkChannel,
                [`& .${tabClasses.root}`]: {
                    py: 1,
                    flex: 1,
                    // transition: '0.3s',
                    fontWeight: 'md',
                    fontSize: 'md',
                    // [`&:not(.${tabClasses.selected}):not(:hover)`]: {
                    //     opacity: 0.7,
                    // },
                },
            })}
        >
        <Box
            border={1}
            borderTop={3}
            borderLeft={3}
            borderRight={3}
            borderBottom={0}
            borderColor="primary.800"
            sx={{
                position: 'fixed',
                bottom: 0,
                left: 0,
                right: 0,
                zIndex: 9999,
                flexGrow: 1,
                maxWidth: 540,
                mx: 'auto',
                // m: -3,
                p: 0.5,
                paddingTop: 0,
                borderTopLeftRadius: '12px',
                borderTopRightRadius: '12px',
                bgcolor: 'common.white',
                // borderColor: 'common.black',
                display: "flex",
                justifyContent: "center",
                alignItems: "center",
            }}
        >
            {!collapsed && (
                <TabList
                    variant="plain"
                    size="sm"
                    disableUnderline
                    sx={{ borderRadius: 'lg', p: 0 }}
                >
                    <Tab
                        disableIndicator
                        orientation="vertical"
                        {...(index === 0 && { color: colors[0] })}
                        sx={{m: 0.5}}
                    >
                        <ListItemDecorator>
                            <PlayArrowIcon />
                        </ListItemDecorator>
                        Play
                    </Tab>
                    <Tab
                        disableIndicator
                        orientation="vertical"
                        {...(index === 1 && { color: colors[1] })}
                        sx={{m: 0.5}}
                    >
                        <ListItemDecorator>
                            <ConnectWithoutContactIcon />
                        </ListItemDecorator>
                        Connect
                    </Tab>
                    <Tab
                        disableIndicator
                        orientation="vertical"
                        {...(index === 2 && { color: colors[2] })}
                        sx={{m: 0.5}}
                    >
                        <ListItemDecorator>
                            <SportsEsportsIcon />
                        </ListItemDecorator>
                        Controllers
                    </Tab>
                    <Tab
                        disableIndicator
                        orientation="vertical"
                        {...(index === 3 && { color: colors[3] })}
                        sx={{m: 0.5}}
                    >
                        <ListItemDecorator>
                            <ComputerIcon />
                        </ListItemDecorator>
                        Host
                    </Tab>
                    <Tab
                        disableIndicator
                        orientation="vertical"
                        {...(index === 4 && { color: colors[4] })}
                        sx={{m: 0.5}}
                    >
                        <ListItemDecorator>
                            <SettingsIcon />
                        </ListItemDecorator>
                        Admin
                    </Tab>
                </TabList>
            )}

            <Button fullWidth={true} size='sm' variant="plain" onClick={handleCollapseToggle}>
                {collapsed ? <ExpandLessIcon /> : <ExpandMoreIcon />}
            </Button>
        </Box>

            <TabPanel 
                value={0}
                sx={{p: 0, m: 0}}
            >
                <PlayTab streamHook={streamHook}></PlayTab>
            </TabPanel>
            <TabPanel 
                value={1}
                sx={{p: 0, m: 0}}
            >
                <ClientPlayer streamHook={streamHook}></ClientPlayer>
            </TabPanel>
            <TabPanel 
                value={2}
            >
                <ControllersTab streamHook={streamHook}></ControllersTab>
            </TabPanel>
            <TabPanel 
                value={3}
                sx={{p: 0, m: 0}}
            >
                <HostMedia castHook={castHook}></HostMedia>
            </TabPanel>
            <TabPanel 
                value={4}
            >
                <AdminPage streamHook={streamHook}></AdminPage>
            </TabPanel>

        </Tabs>
    );
}
