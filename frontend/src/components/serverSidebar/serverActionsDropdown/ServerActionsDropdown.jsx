import React, { useState, useRef, useEffect } from 'react';
import { MoreVertical } from 'lucide-react';

import AddUsersModal from './AddUsersModal';
import ConfirmDeleteModal from './ConfirmDeleteModal';

import { api } from '../../../api/core';

const ServerActionsDropdown = ({ serverId }) => {
  const [isMenuOpen, setIsMenuOpen] = useState(false);
  const [showAddUsersModal, setShowAddUsersModal] = useState(false);
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const dropdownRef = useRef(null);

  const onAddUsers = (username) => {
    api.post(`/servers/${serverId}/users`, { username })
      .then(() => {
        console.log(`User ${username} added to server ${serverId}`);
      })
      .catch((error) => {
        console.error('Error adding user to server:', error);
      });
    setShowAddUsersModal(false);
    setIsMenuOpen(false);
  };

  const onDeleteServer = () => {
    api.delete(`/servers/${serverId}`)
      .then(() => {
        console.log(`Server ${serverId} deleted successfully`);
      })
      .catch((error) => {
        console.error('Error deleting server:', error);
      });
  };
  // Close dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (event) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target)) {
        setIsMenuOpen(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  return (
    <div className="relative" ref={dropdownRef}>
      <button
        onClick={() => setIsMenuOpen((prev) => !prev)}
        className="p-1 rounded hover:bg-gray-200"
      >
        <MoreVertical size={18} />
      </button>

      {isMenuOpen && (
        <div className="absolute right-0 mt-1 w-40 bg-white border border-gray-200 rounded shadow-lg z-10">
          <button
            onClick={() => {
              setIsMenuOpen(false);
              setShowAddUsersModal(true);
            }}
            className="block w-full text-left px-4 py-2 text-sm hover:bg-gray-100"
          >
            Add Users
          </button>
          <button
            onClick={() => {
              setIsMenuOpen(false);
              setShowDeleteModal(true);
            }}
            className="block w-full text-left px-4 py-2 text-sm text-red-600 hover:bg-gray-100"
          >
            Delete Server
          </button>
        </div>
      )}

      {/* Modals */}
      {showAddUsersModal && (
        <AddUsersModal
          onClose={() => setShowAddUsersModal(false)}
          onSubmit={onAddUsers}
        />
      )}

      {showDeleteModal && (
        <ConfirmDeleteModal
          onClose={() => setShowDeleteModal(false)}
          onConfirm={onDeleteServer}
        />
      )}
    </div>
  );
};

export default ServerActionsDropdown;
