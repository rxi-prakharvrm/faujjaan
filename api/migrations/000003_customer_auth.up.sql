-- Add customer role
INSERT INTO roles (key) VALUES ('customer') ON CONFLICT (key) DO NOTHING;

-- Add user_id to carts
ALTER TABLE carts ADD COLUMN user_id UUID REFERENCES users(id) ON DELETE SET NULL;
CREATE INDEX carts_user_id_idx ON carts(user_id);
