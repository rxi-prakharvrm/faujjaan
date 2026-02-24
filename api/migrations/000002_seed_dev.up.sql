-- Seed roles
INSERT INTO roles (key) VALUES
  ('admin'),
  ('customer')
ON CONFLICT (key) DO NOTHING;

-- Seed bootstrap admin
-- Password hash uses bcrypt via pgcrypto's crypt(); value is different per run.
WITH admin_user AS (
  INSERT INTO users (email, password_hash, is_active)
  VALUES ('admin@example.com', crypt('admin12345', gen_salt('bf', 10)), TRUE)
  ON CONFLICT (email) DO UPDATE SET is_active = EXCLUDED.is_active
  RETURNING id
)
INSERT INTO user_roles (user_id, role_id)
SELECT au.id, r.id
FROM admin_user au
JOIN roles r ON r.key = 'admin'
ON CONFLICT DO NOTHING;

-- Categories
INSERT INTO categories (name, slug) VALUES
  ('T-Shirts', 't-shirts'),
  ('Jeans', 'jeans')
ON CONFLICT (slug) DO NOTHING;

-- Products
WITH p AS (
  INSERT INTO products (slug, name, description, status)
  VALUES
    ('classic-tee', 'Classic Tee', 'Everyday cotton tee.', 'active'),
    ('slim-jean', 'Slim Jean', 'Comfort stretch denim.', 'active')
  ON CONFLICT (slug) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    status = EXCLUDED.status,
    updated_at = now()
  RETURNING id, slug, name
),
variants AS (
  INSERT INTO product_variants (product_id, sku, title, size, color, price_inr, compare_at_price_inr)
  SELECT
    p.id,
    CASE p.slug
      WHEN 'classic-tee' THEN 'TEE-BLK-M'
      WHEN 'slim-jean' THEN 'JEAN-BLU-32'
    END,
    CASE p.slug
      WHEN 'classic-tee' THEN 'Black / M'
      WHEN 'slim-jean' THEN 'Blue / 32'
    END,
    CASE p.slug
      WHEN 'classic-tee' THEN 'M'
      WHEN 'slim-jean' THEN '32'
    END,
    CASE p.slug
      WHEN 'classic-tee' THEN 'Black'
      WHEN 'slim-jean' THEN 'Blue'
    END,
    CASE p.slug
      WHEN 'classic-tee' THEN 79900
      WHEN 'slim-jean' THEN 199900
    END,
    NULL
  FROM p
  ON CONFLICT (sku) DO UPDATE SET
    title = EXCLUDED.title,
    size = EXCLUDED.size,
    color = EXCLUDED.color,
    price_inr = EXCLUDED.price_inr,
    compare_at_price_inr = EXCLUDED.compare_at_price_inr,
    updated_at = now()
  RETURNING id
)
INSERT INTO inventory (variant_id, on_hand, reserved)
SELECT v.id, 50, 0
FROM variants v
ON CONFLICT (variant_id) DO UPDATE SET on_hand = EXCLUDED.on_hand, reserved = EXCLUDED.reserved, updated_at = now();

