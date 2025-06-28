import React, { useEffect, useRef, useState } from 'react';
import { useWebSocket } from '../../contexts/WebSocketContext'; // adjust path if needed

const Chat = ({ channel }) => {
  const [messages, setMessages] = useState([]);
  const [input, setInput] = useState('');
  const bottomRef = useRef(null);
  const wsRef = useWebSocket();

  useEffect(() => {
    if (!channel?.id || !wsRef?.current) return;
    const ws = wsRef.current;

    const handleMessage = (event) => {
      try {
        const msg = JSON.parse(event.data);
        if (msg.channel_id === channel.id) {
          setMessages((prev) => [...prev, msg]);
        }
      } catch (err) {
        console.error('Invalid message format', event.data);
      }
    };

    ws.addEventListener('message', handleMessage);
    return () => ws.removeEventListener('message', handleMessage);
  }, [channel.id, wsRef]);

  const sendMessage = () => {
    if (!input.trim() || !wsRef?.current) return;
    const msg = { content: input, channel_id: channel.id, 'server_id': channel.server_id };
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
