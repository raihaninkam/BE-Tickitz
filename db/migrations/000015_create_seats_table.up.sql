-- public.seats definition

-- Drop table

-- DROP TABLE public.seats;
CREATE SEQUENCE seats_id_seq INCREMENT 1 START 0 MINVALUE 0;
CREATE TABLE public.seats (
	id int4 DEFAULT nextval('"seats_id_seq"'::regclass) NOT NULL,
	seats_map varchar(50) NULL,
	orders_id int4 NULL,
	CONSTRAINT "seats_pkey" PRIMARY KEY (id)
);


-- public.seats foreign keys

ALTER TABLE public.seats ADD CONSTRAINT "seats_orders_id_fkey" FOREIGN KEY (orders_id) REFERENCES public.orders(id);