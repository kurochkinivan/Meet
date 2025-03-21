CREATE EXTENSION postgis;

CREATE TABLE IF NOT EXISTS users (
    id UUID DEFAULT gen_random_uuid() NOT NULL,
    name TEXT NOT NULL,
    birthday DATE NOT NULL,
    sex TEXT NOT NULL,
    phone TEXT UNIQUE NOT NULL,
    password TEXT,
    location GEOGRAPHY(Point, 4326),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
    PRIMARY KEY (id),
    CONSTRAINT sex_check CHECK (sex IN ('male', 'female'))
);

CREATE TABLE IF NOT EXISTS photos (
    id INT GENERATED ALWAYS AS IDENTITY,
    user_id UUID NOT NULL,
    object_key TEXT NOT NULL,
    url TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
    PRIMARY KEY(id),
    CONSTRAINT fk_user_id FOREIGN KEY (user_id) REFERENCES users (id)
        ON UPDATE CASCADE ON DELETE CASCADE
);