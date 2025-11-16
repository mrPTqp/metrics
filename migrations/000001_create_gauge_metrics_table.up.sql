CREATE TABLE gauges (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    value DOUBLE PRECISION NOT NULL
);

CREATE UNIQUE INDEX idx_gauges_name ON gauges(name);
