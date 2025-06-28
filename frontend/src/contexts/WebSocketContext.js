// WebSocketContext.js
import React, { createContext, useContext, useEffect, useRef } from 'react';
import ReconnectingWebSocket from 'reconnecting-websocket';

const WebSocketContext = createContext(null);

export const WebSocketProvider = ({ children }) => {
  const token = localStorage.getItem('token');
  const wsRef = useRef(null);

  useEffect(() => {
    if (!token) return;

    const socketUrl = `ws://localhost:8080/ws?token=${token}`;
    const ws = new ReconnectingWebSocket(socketUrl);
    wsRef.current = ws;

    ws.addEventListener('open', () => console.log('WebSocket connected'));
    ws.addEventListener('close', () => console.log('WebSocket disconnected'));
    ws.addEventListener('error', (err) => console.error('WebSocket error:', err));

    return () => ws.close();
  }, [token]);

  return (
    <WebSocketContext.Provider value={wsRef}>
      {children}
    </WebSocketContext.Provider>
  );
};

export const useWebSocket = () => useContext(WebSocketContext);
