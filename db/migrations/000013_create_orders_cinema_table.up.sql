-- public.orders_cinema definition

-- Drop table

-- DROP TABLE public.orders_cinema;

CREATE TABLE public.orders_cinema (
	orders_id int4 NULL,
	cinema_id int4 NULL
);


-- public.orders_cinema foreign keys

ALTER TABLE public.orders_cinema ADD CONSTRAINT "orders_cinema_cinema_id_fkey" FOREIGN KEY (cinema_id) REFERENCES public.cinemas(id);
ALTER TABLE public.orders_cinema ADD CONSTRAINT "orders_cinema_orders_id_fkey" FOREIGN KEY (orders_id) REFERENCES public.orders(id);