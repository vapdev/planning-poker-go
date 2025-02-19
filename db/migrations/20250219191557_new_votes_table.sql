-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS votes (
    room_id INTEGER,
    user_id INTEGER,
    vote varchar(5),
    issue_id INTEGER,
    FOREIGN KEY(room_id) REFERENCES rooms(id),
    FOREIGN KEY(user_id) REFERENCES users(id),
    FOREIGN KEY(issue_id) REFERENCES issues(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
); 
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS votes;

-- +goose StatementEnd