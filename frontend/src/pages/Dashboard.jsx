import React, { useContext } from 'react';
import { AuthContext } from '../contexts/AuthContext';

const Dashboard = () => {
  const { user } = useContext(AuthContext);

  return (
    <div className="min-h-screen bg-gray-100 flex items-center justify-center p-4">
      <div className="bg-white rounded-lg shadow-xl p-8 max-w-md w-full transform transition-all hover:scale-105">
        <h1 className="text-3xl font-bold text-gray-800 mb-6 text-center">
          Dashboard
        </h1>
        <div className="space-y-4">
          <p className="text-lg text-gray-600">
            <span className="font-semibold text-gray-800">Welcome,</span> {user.Username}!
          </p>
          <p className="text-lg text-gray-600">
            <span className="font-semibold text-gray-800">Email:</span> {user.Email}
          </p>
        </div>
      </div>
    </div>
  );
};

export default Dashboard;