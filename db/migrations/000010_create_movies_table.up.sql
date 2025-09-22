-- public.movies definition

-- Drop table

-- DROP TABLE public.movies;
CREATE SEQUENCE movies_id_seq INCREMENT 1 START 0 MINVALUE 0;
CREATE TABLE public.movies (
	id int4 DEFAULT nextval('"movies_id_seq"'::regclass) NOT NULL,
	title varchar(255) NOT NULL,
	synopsis text NULL,
	duration_minutes int4 NULL,
	release_date date NULL,
	poster_image varchar(255) NULL,
	directors_id int4 NULL,
	rating float8 NULL,
	bg_path varchar(255) NULL,
	is_deleted bool DEFAULT false NULL,
	deleted_at timestamp NULL,
	CONSTRAINT "movies_pkey" PRIMARY KEY (id)
);


-- public.movies foreign keys

ALTER TABLE public.movies ADD CONSTRAINT "movies_directors_id_fkey" FOREIGN KEY (directors_id) REFERENCES public.directors(id);