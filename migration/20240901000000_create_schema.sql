-- +goose Up
-- +goose StatementBegin
CREATE TABLE links
(
    link_id serial PRIMARY KEY,
    url     text NOT NULL,
    key     text NOT NULL
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
