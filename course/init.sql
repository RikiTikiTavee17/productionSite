CREATE TABLE note (
    id BIGINT PRIMARY KEY,
    name TEXT NOT NULL,
    price BIGINT,
    description TEXT,
    compositon TEXT,
    author BIGINT,
    photo_url TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP
);

CREATE TABLE persons (
    id BIGINT NOT NULL,
    login TEXT PRIMARY KEY,
    password TEXT NOT NULL,
    position TEXT NOT NULL DEFAULT 'user'
);
