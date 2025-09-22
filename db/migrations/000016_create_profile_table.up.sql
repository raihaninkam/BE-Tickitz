-- public.profile definition

-- Drop table

-- DROP TABLE public.profile;

CREATE TABLE public.profile (
	id SERIAL NOT NULL,
	first_name varchar(50) NOT NULL,
	last_name varchar(50) NOT NULL,
	phone_number varchar(20) NOT NULL,
	profile_picture varchar(100) NULL,
	created_at timestamp DEFAULT now() NULL,
	updated_at timestamp DEFAULT now() NULL,
	CONSTRAINT "profile_pkey" PRIMARY KEY (id)
);


-- public.profile foreign keys

ALTER TABLE public.profile ADD CONSTRAINT "profile_id_fkey" FOREIGN KEY (id) REFERENCES public.users(id);