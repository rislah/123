CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username TEXT NOT NULL,
    password_hash NOT NULL
    role_id INTEGER REFERENCES roles(id) DEFAULT 0
)


CREATE TABLE roles (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL
)
