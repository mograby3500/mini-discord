import React, { useContext } from 'react';
import { AuthContext } from '../contexts/AuthContext';

const Dashboard = () => {
  const { user } = useContext(AuthContext);
    return (
        <div>
            <h1>Dashboard</h1>
            <div>
                <p>Welcome, {user.username}!</p>
                <p>Email: {user.email}</p>
            </div>
        </div>
    );
};

export default Dashboard;
