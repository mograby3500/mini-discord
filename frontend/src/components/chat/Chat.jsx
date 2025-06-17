import React, { useEffect, useRef, useState } from 'react';
import ReconnectingWebSocket from 'reconnecting-websocket';

const Chat = ({ channel }) => {
  const [messages, setMessages] = useState([]);
  const [input, setInput] = useState('');
  const wsRef = useRef(null);
  const bottomRef = useRef(null);
  const token = localStorage.getItem('token');

  useEffect(() => {
    if (!channel?.id || !token) return;

    const socketUrl = `ws://localhost:8080/ws?channel_id=${channel.id}&token=${token}`;
    const ws = new ReconnectingWebSocket(socketUrl);

    wsRef.current = ws;

    ws.addEventListener('open', () => {
      console.log('Connected to chat');
    });

    ws.addEventListener('message', (event) => {
      try {
        console.log(event.data);
        const msg = JSON.parse(event.data);
        setMessages((prev) => [...prev, msg]);
      } catch (err) {
        console.error('Invalid message format', event.data);
      }
    });

    ws.addEventListener('error', (err) => {
      console.error('WebSocket error:', err);
    });

    ws.addEventListener('close', () => {
      console.log('WebSocket disconnected');
    });

    return () => {
      ws.close();
    };
  }, [channel.id, token]);

  const sendMessage = () => {
    if (!input.trim() || !wsRef.current) return;
    const msg = { content: input };
    wsRef.current.send(JSON.stringify(msg));
    setInput('');
  };

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  return (
    <div className="flex flex-col flex-grow bg-white rounded-lg shadow-md overflow-hidden">
      <div className="p-4 border-b font-semibold text-lg text-gray-800">
        #{channel.name}
      </div>
      <div className="flex-grow overflow-y-auto p-4 space-y-2">
        {messages.map((m, i) => (
          <div key={i} className="bg-gray-100 p-2 rounded">
            <span className="font-semibold">User {m.user_id}:</span> {m.content}
          </div>
        ))}
        <div ref={bottomRef} />
      </div>
      <div className="p-4 border-t flex space-x-2">
        <input
          className="flex-grow p-2 border rounded"
          value={input}
          placeholder="Type a message..."
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && sendMessage()}
        />
        <button
          className="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600"
          onClick={sendMessage}
        >
          Send
        </button>
      </div>
    </div>
  );
};

export default Chat;
