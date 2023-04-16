BEGIN;

ALTER TABLE user_resetpasswords
    DROP CONSTRAINT user_resetpasswords_user_id_fkey;

ALTER TABLE user_resetpasswords
    ADD CONSTRAINT user_resetpasswords_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES users(id);

COMMIT;
