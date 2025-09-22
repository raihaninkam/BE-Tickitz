-- public.payment definition

-- Drop table

-- DROP TABLE public.payment;
CREATE SEQUENCE payment_id_seq INCREMENT 1 START 0 MINVALUE 0;
CREATE TABLE public.payment (
	id int4 DEFAULT nextval('"payment_id_seq"'::regclass) NOT NULL,
	"method" varchar(50) NOT NULL,
	CONSTRAINT "payment_pkey" PRIMARY KEY (id)
);