CREATE TABLE counters (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    value BIGINT NOT NULL
);

CREATE UNIQUE INDEX idx_counters_name ON counters(name);
