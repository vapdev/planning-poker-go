-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS room_users (
    room_id INTEGER,
    user_id INTEGER,
    FOREIGN KEY(room_id) REFERENCES rooms(id),
    FOREIGN KEY(user_id) REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
); 
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS room_users;

-- +goose StatementEnd