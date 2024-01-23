CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
    id  UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    name TEXT NOT NULL,
    plan TEXT NULL,
    mail TEXT NULL
);