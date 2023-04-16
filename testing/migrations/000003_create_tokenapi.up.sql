BEGIN;

CREATE TABLE tokenapi_keys(
    id           UUID PRIMARY KEY
    ,user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE
    ,created     TIMESTAMP NOT NULL
    ,deleted     TIMESTAMP
    ,last_used   TIMESTAMP
    ,description VARCHAR(512)
);
CREATE INDEX idx_tokenapi_keys_user_id ON tokenapi_keys(user_id);

COMMIT;
