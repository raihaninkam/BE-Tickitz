-- public.users definition

-- Drop table

-- DROP TABLE public.users;
CREATE SEQUENCE users_id_seq INCREMENT 1 START 0 MINVALUE 0;
CREATE TABLE public.users (
	id int4 DEFAULT nextval('"users_id_seq"'::regclass) NOT NULL,
	email varchar(50) NOT NULL,
	"password" text NOT NULL,
	poin float8 DEFAULT 0 NULL,
	"role" varchar(20) NULL,
	CONSTRAINT "users_pkey" PRIMARY KEY (id)
);