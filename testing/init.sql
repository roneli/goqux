DROP TABLE IF EXISTS "insert_posts";
DROP TABLE IF EXISTS "users";
DROP TABLE IF EXISTS "select_users";


CREATE TABLE IF NOT EXISTS "users"
(
    "id"         SERIAL PRIMARY KEY,
    "username"   VARCHAR(255) NOT NULL,
    "password"   VARCHAR(255) NOT NULL,
    "email"      VARCHAR(255) NOT NULL,
    "created_at" TIMESTAMP    NOT NULL DEFAULT NOW(),
    "updated_at" TIMESTAMP    NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS "select_users"
(
    "id"         SERIAL PRIMARY KEY,
    "username"   VARCHAR(255) NOT NULL,
    "password"   VARCHAR(255) NOT NULL,
    "email"      VARCHAR(255) NOT NULL,
    "created_at" TIMESTAMP    NOT NULL DEFAULT NOW(),
    "updated_at" TIMESTAMP    NOT NULL DEFAULT NOW()
);

INSERT INTO "select_users" ("username", "password", "email")
VALUES ('admin', 'admin', 'admin@acme.com'),
       ('user', 'user', 'user@acme.com');

-- posts table is used for insert testing
CREATE TABLE IF NOT EXISTS "insert_posts"
(
    "id"         SERIAL PRIMARY KEY,
    "title"      VARCHAR(255) NOT NULL,
    "content"    TEXT         NOT NULL,
    "created_at" TIMESTAMP    NOT NULL DEFAULT NOW(),
    "updated_at" TIMESTAMP    NOT NULL DEFAULT NOW(),
    "user_id"    INTEGER      NOT NULL REFERENCES "users" ("id")
);