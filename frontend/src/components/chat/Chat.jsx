import React, { useEffect, useRef, useState } from 'react';
import { useWebSocket } from '../../contexts/WebSocketContext'; // adjust path if needed
import { api } from '../../api/core';
import Message  from './Message';

const Chat = ({ channel }) => {
  const [messages, setMessages] = useState([]);
  const [loading, setLoading] = useState(false);
  const [hasMore, setHasMore] = useState(true);
  const [input, setInput] = useState('');
  const bottomRef = useRef(null);
  const wsRef = useWebSocket();
  const messagesContainerRef = useRef(null);
  const firstMessageRef = useRef(null);
  const [firstLoadDone, setFirstLoadDone] = useState(false);


  useEffect(() => {
    let isMounted = true;

    const fetchMessages = async () => {
      try {
        const response = await api.get(`/messages/${channel.id}?limit=50`);
        if (isMounted) {
          setMessages((prev) => {
            const newMessages = response.data.reverse();
            const updated = [...newMessages, ...prev];
            setTimeout(() => {
              scrollToBottom();
            }, 0);
            return updated;
          });
          setTimeout(() => {
            setFirstLoadDone(true);
          }, 500);
        }
      } catch (error) {
        console.error('Error fetching initial messages:', error);
      }
    };

    if (channel?.id) fetchMessages();

    return () => {
      isMounted = false;
    };
  }, [channel.id]);

  // WebSocket message handling
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
    const msg = { content: input, channel_id: channel.id, server_id: channel.server_id };
    wsRef.current.send(JSON.stringify(msg));
    setInput('');
  };

  useEffect(() => {
    if (!firstLoadDone || loading) return;
    const observer = new IntersectionObserver(
      async (entries) => {
        if (entries[0].isIntersecting && hasMore) {
          setLoading(true);
          try {
            const oldestMessageId = messages.length > 0 ? messages[0].id : null;
            const response = await api.get(
              `/messages/${channel.id}?before=${oldestMessageId}&limit=50`
            );
            const newMessages = response.data.reverse();

            if (newMessages.length > 0) {
              const container = messagesContainerRef.current;
              const previousScrollHeight = container.scrollHeight;

              setMessages((prev) => [...newMessages, ...prev]);

              requestAnimationFrame(() => {
                const newScrollHeight = container.scrollHeight;
                container.scrollTop = newScrollHeight - previousScrollHeight;
              });
            } else {
              setHasMore(false);
            }
          } catch (error) {
            console.error('Error fetching more messages:', error);
          }
          setLoading(false);
        }
      },
      {
        root: messagesContainerRef.current,
        threshold: 0.1,
      }
    );

    const current = firstMessageRef.current;
    if (current) observer.observe(current);

    return () => {
      if (current) observer.unobserve(current);
    };
  }, [messages, channel.id, loading, hasMore, firstLoadDone]);


  const scrollToBottom = () => {
    if (bottomRef.current) {
      bottomRef.current.scrollIntoView();
    }
  };

  return (
    <div className="flex flex-col flex-grow bg-white rounded-lg shadow-md overflow-hidden">
      <div className="p-4 border-b font-semibold text-lg text-gray-800">
        #{channel.name}
      </div>
      <div
        ref={messagesContainerRef}
        className="flex-grow overflow-y-auto p-4 space-y-2"
      >
        {messages.map((m, index) => (
          <Message
            key={m.id}
            message={m}
            ref={index === 0 ? firstMessageRef : null}
          />
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
