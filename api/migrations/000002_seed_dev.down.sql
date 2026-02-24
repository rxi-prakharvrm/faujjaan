DELETE FROM inventory;
DELETE FROM product_variants;
DELETE FROM product_categories;
DELETE FROM products;
DELETE FROM categories;
DELETE FROM user_roles;
DELETE FROM users WHERE email = 'admin@example.com';
DELETE FROM roles WHERE key IN ('admin','customer');

