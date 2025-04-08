CREATE TABLE notes(
    id   SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    title VARCHAR,
    content VARCHAR,
    archived BOOLEAN NOT NULL
);