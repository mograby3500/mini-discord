import React, { useState } from 'react';

const CreateServerModal = ({ onClose, onSubmit }) => {
  const [serverName, setServerName] = useState('');

  return (
    <div className="fixed inset-0 bg-black bg-opacity-40 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg shadow-lg p-6 w-96">
        <h2 className="text-lg font-bold mb-4">Create Server</h2>

        <input
          type="text"
          placeholder="Server name"
          value={serverName}
          onChange={(e) => setServerName(e.target.value)}
          className="border border-gray-300 rounded w-full p-2 mb-4"
        />

        <div className="flex justify-end gap-2">
          <button
            onClick={onClose}
            className="px-4 py-2 rounded bg-gray-200 hover:bg-gray-300"
          >
            Cancel
          </button>
          <button
            onClick={() => {
              if (serverName.trim()) {
                onSubmit(serverName.trim());
                onClose();
              }
            }}
            className="px-4 py-2 rounded bg-blue-500 text-white hover:bg-blue-600"
          >
            Create
          </button>
        </div>
      </div>
    </div>
  );
};

export default CreateServerModal;
