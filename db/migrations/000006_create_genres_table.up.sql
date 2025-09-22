-- public.genres definition

-- Drop table

-- DROP TABLE public.genres;

CREATE TABLE public.genres (
	id SERIAL NOT NULL,
	"name" varchar(100),
	CONSTRAINT "genres_name_key" UNIQUE (name),
	CONSTRAINT "genres_pkey" PRIMARY KEY (id)
);