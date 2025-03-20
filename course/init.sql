CREATE TABLE note (
    id INT PRIMARY KEY,
    name TEXT NOT NULL,
    price INT,
    description TEXT,
    composition TEXT,
    author INT,
    photo_url TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP
);

CREATE TABLE persons (
    id INT NOT NULL,
    login TEXT PRIMARY KEY,
    password TEXT NOT NULL,
    position TEXT NOT NULL DEFAULT 'user'
);
