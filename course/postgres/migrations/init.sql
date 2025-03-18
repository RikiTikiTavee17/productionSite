CREATE TABLE note (
    id BIGINT PRIMARY KEY,
    title TEXT NOT NULL,
    content TEXT,
    author BIGINT,
    dead_line TIMESTAMP,
    status BOOLEAN,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP
);

CREATE TABLE persons (
    id BIGINT NOT NULL,
    login TEXT PRIMARY KEY,
    password TEXT NOT NULL
);
