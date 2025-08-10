import React, { useEffect, useState } from 'react';
import { getServers } from '../api/servers/servers';
import Sidebar from '../components/Sidebar';
import ServerSidebar from '../components/serverSidebar/ServerSidebar';
import Chat from '../components/chat/Chat';
import { WebSocketProvider } from '../contexts/WebSocketContext';

const Dashboard = () => {
  const [servers, setServers] = useState([]);
  const [selectedServer, setSelectedServer] = useState(null);
  const [selectedChannel, setSelectedChannel] = useState(null);

  useEffect(() => {
    const fetchServers = async () => {
      try {
        const data = await getServers();
        setServers(data);
        if (data.length > 0) {
          setSelectedServer(data[0]);
        }
      } catch (err) {
        console.error('Failed to fetch servers:', err);
      }
    };

    fetchServers();
  }, []);

  return (
    <WebSocketProvider>
      <div
        className="flex flex-grow"
        style={{ height: 'calc(100vh - 64px)' }}
      >
        <Sidebar
          servers={servers}
          selectedServerId={selectedServer?.id}
          onSelectServer={(server) => {
            setSelectedServer(server);
            setSelectedChannel(server.channels[0]);
          }}
        />
        <ServerSidebar
          server={selectedServer}
          channels={selectedServer?.channels || []}
          selectedChannelId={selectedChannel?.id}
          onSelectChannel={setSelectedChannel}
        />
        <div className="flex-grow flex flex-col">
          {selectedChannel ? (
            <Chat channel={selectedChannel} key={selectedChannel?.id}/>
          ) : (
            <div className="flex flex-grow items-center justify-center text-gray-500">
              Select a channel to start chatting
            </div>
          )}
        </div>
      </div>
    </WebSocketProvider>
  );
};

export default Dashboard;