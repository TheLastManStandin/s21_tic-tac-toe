CREATE TABLE public.client (
    uuid character varying(36) NOT NULL,
    name character varying(30),
    password character varying(60)
);

CREATE TABLE public.games (
    uuid character varying(36) NOT NULL,
    gametable integer[],
    gamestate integer
);

ALTER TABLE ONLY public.client
    ADD CONSTRAINT client_pkey PRIMARY KEY (uuid);

ALTER TABLE ONLY public.games
    ADD CONSTRAINT games_pkey PRIMARY KEY (uuid);
