CREATE TABLE role (
    id   SERIAL PRIMARY KEY,
    name TEXT   NOT NULL
);

INSERT INTO role (name) VALUES ('guest');

CREATE TABLE users (
    id            SERIAL                        PRIMARY KEY,
    user_id       UUID                          NOT     NULL UNIQUE DEFAULT gen_random_uuid(),
    username      TEXT                          NOT     NULL UNIQUE,
    password_hash TEXT                          NOT     NULL,
    created_at    TIMESTAMPTZ                   NOT     NULL DEFAULT now()
);

CREATE TABLE user_role (
    id SERIAL PRIMARY KEY,
    user_id UUID REFERENCES users(user_id) ON DELETE CASCADE ON UPDATE CASCADE NOT NULL UNIQUE,
    role_id INTEGER REFERENCES role(id) ON DELETE CASCADE ON UPDATE CASCADE NOT NULL  DEFAULT 1
);
