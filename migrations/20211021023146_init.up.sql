CREATE TABLE role (
    id   SERIAL PRIMARY KEY,
    name TEXT   NOT NULL
);

CREATE TABLE users (
    id            SERIAL                        PRIMARY KEY,
    user_id       UUID                          NOT     NULL DEFAULT gen_random_uuid(),
    username      TEXT                          NOT     NULL UNIQUE,
    password_hash TEXT                          NOT     NULL,
    created_at    TIMESTAMPTZ                   NOT     NULL DEFAULT now(),
    role_id INTEGER REFERENCES role(id) NOT NULL DEFAULT 1
);

INSERT INTO role (name) VALUES ('guest');
