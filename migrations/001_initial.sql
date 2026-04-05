CREATE TABLE IF NOT EXISTS categories (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS pets (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    category_id BIGINT REFERENCES categories(id),
    status TEXT CHECK (status IN ('available', 'pending', 'sold')) DEFAULT 'available'
);

CREATE TABLE IF NOT EXISTS pet_photo_urls (
    id BIGSERIAL PRIMARY KEY,
    pet_id BIGINT NOT NULL REFERENCES pets(id) ON DELETE CASCADE,
    url TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS tags (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS pet_tags (
    pet_id BIGINT NOT NULL REFERENCES pets(id) ON DELETE CASCADE,
    tag_id BIGINT NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (pet_id, tag_id)
);

CREATE TABLE IF NOT EXISTS orders (
    id BIGSERIAL PRIMARY KEY,
    pet_id BIGINT NOT NULL,
    quantity INTEGER NOT NULL DEFAULT 1,
    ship_date TIMESTAMPTZ,
    status TEXT CHECK (status IN ('placed', 'approved', 'delivered')) DEFAULT 'placed',
    complete BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    username TEXT UNIQUE NOT NULL,
    first_name TEXT,
    last_name TEXT,
    email TEXT,
    password TEXT,
    phone TEXT,
    user_status INTEGER DEFAULT 0
);
