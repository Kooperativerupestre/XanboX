CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL
);


CREATE TABLE tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    made_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    image TEXT NOT NULL,
    environment_prepare_code TEXT[] NOT NULL,
    execution_code TEXT NOT NULL,
    source TEXT NOT NULL,
    maker UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status TEXT NOT NULL CHECK(status IN ('pending', 'failed', 'successful')) DEFAULT 'pending'
);


