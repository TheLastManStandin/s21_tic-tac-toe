BEGIN;

ALTER TABLE games
    DROP COLUMN secondplayer,
    DROP COLUMN firstplayer;

COMMIT;
