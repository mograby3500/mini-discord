import React, { useEffect } from 'react';
import ServerActionsDropdown from './serverActionsDropdown/ServerActionsDropdown';

const ServerSidebar = ({ server, channels, selectedChannelId, onSelectChannel, deleteServer }) => {
  if (!server) return null;
  return (
    <div className="w-80 bg-white shadow-lg rounded-lg p-4 space-y-2">
      <div className="flex items-center justify-between mb-2 border-b pb-2">
        <h1 className="text-lg font-bold text-gray-800">{server?.name}</h1>
        {(server.role === 'owner' || server.role === 'admin') && (
          <ServerActionsDropdown 
            serverId={server?.id} 
            deleteServer={deleteServer} 
          />
        )}
      </div>
      <h2 className="text-lg font-bold text-gray-800 mb-2">Channels</h2>
      <ul className="space-y-1">
        {channels.map((channel) => (
          <li
            key={channel.id}
            onClick={() => onSelectChannel(channel)}
            className={`p-2 rounded cursor-pointer transition 
              ${
                selectedChannelId === channel.id
                  ? 'bg-blue-500 text-white'
                  : 'bg-gray-100 hover:bg-gray-200 text-gray-800'
              }`}
          >
            {channel.name}
          </li>
        ))}
      </ul>
    </div>
  );
};

export default ServerSidebar;
