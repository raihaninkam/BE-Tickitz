-- public.now_showing definition

-- Drop table

-- DROP TABLE public.now_showing;
CREATE SEQUENCE now_showing_id_seq INCREMENT 1 START 0 MINVALUE 0;
CREATE TABLE public.now_showing (
	id int4 DEFAULT nextval('"now_showing_id_seq"'::regclass) NOT NULL,
	"date" date NOT NULL,
	"time" time NOT NULL,
	location_id int4 NULL,
	movie_id int4 NULL,
	CONSTRAINT "now_showing_pkey" PRIMARY KEY (id)
);


-- public.now_showing foreign keys

ALTER TABLE public.now_showing ADD CONSTRAINT "now_showing_location_id_fkey" FOREIGN KEY (location_id) REFERENCES public."location"(id);
ALTER TABLE public.now_showing ADD CONSTRAINT "now_showing_movie_id_fkey" FOREIGN KEY (movie_id) REFERENCES public.movies(id);