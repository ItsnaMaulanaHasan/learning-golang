CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    department TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE
);