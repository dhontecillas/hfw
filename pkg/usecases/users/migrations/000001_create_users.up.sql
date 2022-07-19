BEGIN;

CREATE TABLE users(
    id           UUID PRIMARY KEY
    ,email       VARCHAR(254) UNIQUE
    ,password    VARCHAR(128)
    ,created     TIMESTAMP
);
CREATE INDEX idx_users_email ON users(email);

CREATE TABLE user_registration_requests(
    token           VARCHAR(254) PRIMARY KEY
    ,email          VARCHAR(254) NOT NULL
    ,requested      TIMESTAMP NOT NULL
    ,expires        TIMESTAMP NOT NULL
    ,password       VARCHAR(128)
    ,consumed       TIMESTAMP
);

CREATE TABLE user_resetpasswords(
    token           VARCHAR(254) PRIMARY KEY
    ,user_id        UUID REFERENCES users(id)
    ,requested      TIMESTAMP NOT NULL
    ,expires        TIMESTAMP NOT NULL
    ,consumed       TIMESTAMP
);

COMMIT;
