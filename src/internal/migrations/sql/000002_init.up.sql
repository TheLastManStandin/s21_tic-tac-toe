BEGIN;

ALTER TABLE games
    ADD COLUMN firstplayer UUID REFERENCES client(uuid),
    ADD COLUMN secondplayer UUID REFERENCES client(uuid);

COMMIT;
