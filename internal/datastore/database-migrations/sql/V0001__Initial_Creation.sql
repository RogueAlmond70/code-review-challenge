CREATE TABLE notes(
    id   SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    title VARCHAR,
    content VARCHAR,
    archived BOOLEAN NOT NULL
);

CREATE TABLE users (
    user_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(255) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role VARCHAR(50) DEFAULT 'user',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
