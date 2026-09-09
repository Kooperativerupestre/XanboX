CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL
);


CREATE TABLE tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    made_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    language TEXT NOT NULL,
    source TEXT NOT NULL,
    maker UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE
);


