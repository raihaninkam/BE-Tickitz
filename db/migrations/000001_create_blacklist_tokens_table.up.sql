-- public.blacklist_tokens definition

-- Drop table

-- DROP TABLE public.blacklist_tokens;

CREATE TABLE public.blacklist_tokens (
	id serial4 NOT NULL,
	"token" text NOT NULL,
	expires_at timestamptz NOT NULL,
	created_at timestamptz DEFAULT now() NULL,
	CONSTRAINT blacklist_tokens_pkey PRIMARY KEY (id),
	CONSTRAINT blacklist_tokens_token_key UNIQUE (token)
);
CREATE INDEX idx_blacklist_expires_at ON public.blacklist_tokens USING btree (expires_at);
CREATE INDEX idx_blacklist_token ON public.blacklist_tokens USING btree (token);