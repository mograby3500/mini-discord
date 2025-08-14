import React, { useState } from 'react';
import { Plus } from 'lucide-react';
import CreateServerModal from './CreateServerModal';
import { api } from '../api/core';

const getInitials = (name) => {
  return name
    .split(' ')
    .map((word) => word[0]?.toUpperCase())
    .join('');
};

const Sidebar = ({ servers, addNewServer, selectedServerId, onSelectServer }) => {
  const [showCreateModal, setShowCreateModal] = useState(false);

  const onCreateServer = (name) => {
    api.post('/servers', { name })
    .then((response) => {
      addNewServer(response.data[0]);
    })
    .catch((error) => {
      console.error('Error creating server:', error);
    });
  };

  return (
    <>
      <div className="w-20 bg-white shadow-lg rounded-lg p-2 flex flex-col items-center">
        <div className="flex flex-col items-center space-y-4">
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
        <button
          onClick={() => setShowCreateModal(true)}
          className="w-12 h-12 mt-4 rounded-full flex items-center justify-center 
                    bg-gray-200 text-gray-700 hover:bg-gray-300 
                    transition-colors duration-200"
          title="Create Server"
        >
          <Plus size={20} />
        </button>
      </div>

      {/* Modal */}
      {showCreateModal && (
        <CreateServerModal
          onClose={() => setShowCreateModal(false)}
          onSubmit={(name) => {
            onCreateServer(name);
            setShowCreateModal(false);
          }}
        />
      )}
    </>
  );
};

export default Sidebar;

