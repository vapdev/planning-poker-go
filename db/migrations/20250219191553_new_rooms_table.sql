-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS rooms (
    id SERIAL PRIMARY KEY,
    uuid UUID,
    name varchar(255),
    showCards BOOLEAN DEFAULT FALSE,
    autoShowCards BOOLEAN DEFAULT FALSE,
    deck TEXT,
    admin INTEGER,
    currentIssue INTEGER,
    lastActive TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
); 
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS rooms;

-- +goose StatementEnd