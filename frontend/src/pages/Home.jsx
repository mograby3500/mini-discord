import React from 'react';
import { Link } from 'react-router-dom';

const Home = () => {
  return (
    <div className="min-h-screen bg-gray-900 text-white flex flex-col items-center justify-center p-4">
      <div className="text-center max-w-2xl mx-auto">
        <h1 className="text-5xl md:text-6xl font-bold tracking-tight mb-6 bg-clip-text text-transparent bg-gradient-to-r from-blue-400 to-purple-500">
          Welcome to ConnectSphere
        </h1>
        <p className="text-lg md:text-xl text-gray-300 mb-8">
          Join a vibrant community to chat, collaborate, and share with friends and like-minded people. Your space, your voice, your way.
        </p>
        <div className="flex justify-center gap-4">
          <Link
            to="/register"
            className="bg-blue-600 hover:bg-blue-700 text-white font-semibold py-3 px-6 rounded-lg transition-colors duration-300 shadow-md hover:shadow-lg"
          >
            Get Started
          </Link>
          <Link
            to="/login"
            className="bg-gray-700 hover:bg-gray-600 text-white font-semibold py-3 px-6 rounded-lg transition-colors duration-300 shadow-md hover:shadow-lg"
          >
            Log In
          </Link>
        </div>
      </div>
      <div className="mt-12 grid grid-cols-1 md:grid-cols-3 gap-6 max-w-4xl mx-auto">
        <div className="bg-gray-800 p-6 rounded-lg shadow-lg transform transition-all hover:scale-105">
          <h3 className="text-xl font-semibold text-blue-400 mb-2">Chat Freely</h3>
          <p className="text-gray-300">Connect with friends in real-time with text, voice, and video channels.</p>
        </div>
        <div className="bg-gray-800 p-6 rounded-lg shadow-lg transform transition-all hover:scale-105">
          <h3 className="text-xl font-semibold text-purple-400 mb-2">Build Communities</h3>
          <p className="text-gray-300">Create and manage your own servers to bring people together.</p>
        </div>
        <div className="bg-gray-800 p-6 rounded-lg shadow-lg transform transition-all hover:scale-105">
          <h3 className="text-xl font-semibold text-blue-400 mb-2">Stay Connected</h3>
          <p className="text-gray-300">Access your conversations anytime, anywhere, on any device.</p>
        </div>
      </div>
    </div>
  );
};

export default Home;
