-- public.orders definition

-- Drop table

-- DROP TABLE public.orders;
CREATE SEQUENCE orders_id_seq INCREMENT 1 START 0 MINVALUE 0;
CREATE TABLE public.orders (
	id int4 DEFAULT nextval('"orders_id_seq"'::regclass) NOT NULL,
	users_id int4 NULL,
	price int4 NULL,
	payment_id int4 NULL,
	"isPaid" bool DEFAULT false NULL,
	created_at timestamp DEFAULT now() NULL,
	now_showing_id int4 NULL,
	isorder bool NULL,
	CONSTRAINT "orders_pkey" PRIMARY KEY (id)
);


-- public.orders foreign keys

ALTER TABLE public.orders ADD CONSTRAINT "orders_now_showing_id_fkey" FOREIGN KEY (now_showing_id) REFERENCES public.now_showing(id);
ALTER TABLE public.orders ADD CONSTRAINT "orders_payment_id_fkey" FOREIGN KEY (payment_id) REFERENCES public.payment(id);
ALTER TABLE public.orders ADD CONSTRAINT "orders_users_id_fkey" FOREIGN KEY (users_id) REFERENCES public.users(id);