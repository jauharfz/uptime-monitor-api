CREATE TABLE users
(
    id         SERIAL PRIMARY KEY NOT NULL,
    username   VARCHAR(255)       NOT NULL,
    password   VARCHAR(60)        NOT NULL,
    email      VARCHAR(255)       NOT NULL,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);
CREATE TABLE monitors
(
    id             SERIAL PRIMARY KEY NOT NULL,
    user_id        INT                NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    url            VARCHAR(255)       NOT NULL,
    check_interval INT                NOT NULL,
    created_at     TIMESTAMP DEFAULT now(),
    updated_at     TIMESTAMP DEFAULT now()
);
CREATE TABLE checks
(
    id            BIGSERIAL PRIMARY KEY NOT NULL,
    monitor_id    INT                   NOT NULL REFERENCES monitors (id) ON DELETE CASCADE,
    status_code   INT                   NOT NULL,
    response_time INT                   NOT NULL,
    created_at    TIMESTAMP DEFAULT now(),
    updated_at    TIMESTAMP DEFAULT now()
);

CREATE UNIQUE INDEX unique_lower_email_idx ON users (LOWER(email));
CREATE INDEX monitor_id_idx ON checks (monitor_id);
