-- Схема БД для приложения BookCollect

CREATE TABLE IF NOT EXISTS collections (
                                           id SERIAL PRIMARY KEY,
                                           release_number INT,
                                           release_year   INT,
                                           title          TEXT NOT NULL,
                                           description    TEXT,
                                           cover_image    TEXT,
                                           publication_link TEXT NOT NULL DEFAULT '',
                                           pdf_path       TEXT
);

CREATE TABLE IF NOT EXISTS articles (
                                        id SERIAL PRIMARY KEY,
                                        author    TEXT NOT NULL,
                                        title     TEXT NOT NULL,
                                        email     TEXT NOT NULL,
                                        file_path TEXT NOT NULL,
                                        created_at TIMESTAMP NOT NULL DEFAULT NOW()
    );

CREATE TABLE IF NOT EXISTS administrators (
                                              id SERIAL PRIMARY KEY,
                                              login         TEXT UNIQUE NOT NULL,
                                              password_hash TEXT NOT NULL
);

-- Тестовый админ: login=admin, pass=admin (bcrypt $2a$10$...)
INSERT INTO administrators (login, password_hash)
SELECT 'admin', '$2a$10$H0Yc2g9C1QGQk4t3uDq8Bu0tQ8o4zZpX0M1S1J5P1t0cQjz5b5C5K'
    WHERE NOT EXISTS (SELECT 1 FROM administrators WHERE login='admin');
