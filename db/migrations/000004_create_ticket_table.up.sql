-- public.ticket definition

-- Drop table

-- DROP TABLE public.ticket;
CREATE SEQUENCE ticket_id_seq INCREMENT 1 START 0 MINVALUE 0;
CREATE TABLE public.ticket (
	id int4 DEFAULT nextval('"ticket_id_seq"'::regclass) NOT NULL,
	qr_code varchar(50) NULL,
	CONSTRAINT "ticket_pkey" PRIMARY KEY (id),
	CONSTRAINT "ticket_qr_code_key" UNIQUE (qr_code)
);