-- public."location" definition

-- Drop table

-- DROP TABLE public."location";

CREATE TABLE public."location" (
	id serial NOT NULL,
	"name" varchar(50) NOT NULL,
	CONSTRAINT "location_pkey" PRIMARY KEY (id)
);