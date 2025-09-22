-- public.orders_ticket definition

-- Drop table

-- DROP TABLE public.orders_ticket;

CREATE TABLE public.orders_ticket (
	orders_id int4 NULL,
	ticket_id int4 NULL
);


-- public.orders_ticket foreign keys

ALTER TABLE public.orders_ticket ADD CONSTRAINT "orders_ticket_orders_id_fkey" FOREIGN KEY (orders_id) REFERENCES public.orders(id);
ALTER TABLE public.orders_ticket ADD CONSTRAINT "orders_ticket_ticket_id_fkey" FOREIGN KEY (ticket_id) REFERENCES public.ticket(id);