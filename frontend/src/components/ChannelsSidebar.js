import React from 'react';

const ChannelsSidebar = ({ channels, selectedChannelId, onSelectChannel }) => {
  return (
    <div className="w-48 bg-white shadow-lg rounded-lg p-4 space-y-2">
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

export default ChannelsSidebar;
