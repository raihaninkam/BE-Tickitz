-- public.directors definition

-- Drop table

-- DROP TABLE public.directors;

CREATE TABLE public.directors (
	id SERIAL NOT NULL,
	"name" varchar(50) NOT NULL,
	CONSTRAINT "directors_pkey" PRIMARY KEY (id)
);