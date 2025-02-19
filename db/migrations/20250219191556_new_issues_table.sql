-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS issues (
    id SERIAL PRIMARY KEY,
    room_id INTEGER,
    uuid UUID,
    title varchar(255),
    description TEXT,
    link TEXT,
    sequence INTEGER,
    FOREIGN KEY(room_id) REFERENCES rooms(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS issues;

-- +goose StatementEnd