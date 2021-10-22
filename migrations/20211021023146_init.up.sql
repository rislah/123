CREATE TABLE role
(
    id   SERIAL PRIMARY KEY,
    name TEXT   NOT     NULL
);

INSERT INTO roles (name) VALUES ('guest');

CREATE TABLE user
(
    id            SERIAL                        PRIMARY KEY,
    user_id       UUID                          NOT     NULL DEFAULT gen_random_uuid(),
    username      TEXT                          NOT     NULL UNIQUE,
    password_hash TEXT                          NOT     NULL,
    created_at    TIMESTAMPTZ                   NOT     NULL DEFAULT now()
);

CREATE TABLE user_role 
(
    id SERIAL PRIMARY KEY,
    role_id INTEGER REFERENCES role(id) NOT NULL
    user_id INTEGER REFERENCES user(id)  NOT NULL
)



/*
roles => user, dev, admin
permissions => 

*/

