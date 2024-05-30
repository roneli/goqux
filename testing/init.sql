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
-- create a table with two columns one is a serial number and another is a random number, we will then
-- generate a series and insert 100 rows into the table
CREATE TABLE IF NOT EXISTS "random_numbers"
(
    "id"         SERIAL PRIMARY KEY,
    "number"     INTEGER      NOT NULL
);

-- Insert 100 rows into the random_numbers table
DO $$
BEGIN
    FOR i IN 1..100 LOOP
        INSERT INTO "random_numbers" ("number") VALUES (floor(random() * 1000));
    END LOOP;
END $$;