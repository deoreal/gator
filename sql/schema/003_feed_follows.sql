-- +goose Up
CREATE TABLE feed_follows (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  user_id UUID NOT NULL,
  feed_id INTEGER NOT NULL,
  CONSTRAINT fk_user
    FOREIGN KEY(user_id)
    REFERENCES users(id)
    ON DELETE CASCADE,
  CONSTRAINT fk_feed
    FOREIGN KEY(feed_id)
    REFERENCES feeds(id)
    ON DELETE CASCADE,
  UNIQUE (user_id, feed_id)
);

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE 'plpgsql';
-- +goose StatementEnd

CREATE TRIGGER update_feed_follows_updated_at_column
    BEFORE UPDATE ON feed_follows
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- +goose Down
DROP TRIGGER IF EXISTS update_feed_follows_updated_at_column ON feed_follows;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP TABLE feed_follows;
