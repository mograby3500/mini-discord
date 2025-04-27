CREATE TABLE messages (
    id SERIAL PRIMARY KEY,
    channel_id INT REFERENCES channels(id),
    user_id INT REFERENCES users(id),
    content TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
