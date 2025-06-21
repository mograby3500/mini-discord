import React from 'react';

const getInitials = (name) => {
  return name
    .split(' ')
    .map((word) => word[0]?.toUpperCase())
    .join('');
};

const ServersSidebar = ({ servers, selectedServerId, onSelectServer }) => {
  return (
    <div className="w-20 bg-white shadow-lg rounded-lg p-2 flex flex-col items-center space-y-4">
      {servers.map((server) => (
        <div
          key={server.id}
          title={server.name}
          onClick={() => onSelectServer(server)}
          className={`w-12 h-12 rounded-full flex items-center justify-center font-bold text-sm cursor-pointer transition
            ${selectedServerId === server.id
              ? 'bg-blue-500 text-white'
              : 'bg-gray-300 text-gray-800 hover:bg-gray-400'}
          `}
        >
          {getInitials(server.name)}
        </div>
      ))}
    </div>
  );
};

export default ServersSidebar;

