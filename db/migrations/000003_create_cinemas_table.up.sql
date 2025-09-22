-- public.cinemas definition

-- Drop table

-- DROP TABLE public.cinemas;

CREATE SEQUENCE cinemas_id_seq INCREMENT 1 START 0 MINVALUE 0;
CREATE TABLE public.cinemas (
	id int4 DEFAULT nextval('"cinemas_id_seq"'::regclass) NOT NULL,
	cinema_name varchar(50) NOT NULL,
	created_at timestamp DEFAULT now() NULL,
	CONSTRAINT "cinemas_pkey" PRIMARY KEY (id)
);