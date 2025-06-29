import { forwardRef } from 'react';

const Message = forwardRef (({ message }, ref) => {
  const { user_name, content, created_at } = message;

  const initial = user_name?.charAt(0)?.toUpperCase() || '?';

  const createdDate = new Date(created_at);
  const now = new Date();

  const isToday =
    createdDate.getDate() === now.getDate() &&
    createdDate.getMonth() === now.getMonth() &&
    createdDate.getFullYear() === now.getFullYear();

  const timeString = createdDate.toLocaleTimeString([], {
    hour: '2-digit',
    minute: '2-digit',
  });

  const dateString = createdDate.toLocaleDateString();

  const displayTime = isToday ? timeString : `${dateString} ${timeString}`;

  return (
    <div ref={ref} className="bg-gray-100 p-2 rounded">
      <div className="flex items-center space-x-2 text-sm text-gray-700">
        <div className="w-8 h-8 flex items-center justify-center bg-blue-500 text-white rounded-full font-bold">
          {initial}
        </div>
        <span className="font-semibold">{user_name}</span>
        <span className="text-xs text-gray-500">{displayTime}</span>
      </div>
      <div className="text-gray-800">{content}</div>
    </div>
  );
});

export default Message;
