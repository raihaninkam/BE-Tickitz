-- public.casts definition

-- Drop table

-- DROP TABLE public.casts;

CREATE SEQUENCE casts_id_seq INCREMENT 1 START 0 MINVALUE 0;
CREATE TABLE public.casts (
	id int4 DEFAULT nextval('"casts_id_seq"'::regclass) NOT NULL,
	"name" varchar(50) NOT NULL,
	CONSTRAINT "casts_pkey" PRIMARY KEY (id)
);