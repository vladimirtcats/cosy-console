INSERT INTO users (id, email, name) VALUES
    (1, 'ada@example.com', 'Ada'),
    (2, 'bob@example.com', 'Bob'),
    (3, 'cleo@example.com', 'Cleo')
ON CONFLICT (email) DO NOTHING;

INSERT INTO items (id, sku, title, price_cents, active) VALUES
    ('00000000-0000-0000-0000-0000000000a1', 'MUG-01', 'Ceramic mug', 990, true),
    ('00000000-0000-0000-0000-0000000000a2', 'MUG-02', 'Enamel mug', 690, true),
    ('00000000-0000-0000-0000-0000000000a3', 'TEE-01', 'Go gopher tee', 1490, true),
    ('00000000-0000-0000-0000-0000000000a4', 'TEE-02', 'Vim enthusiast tee', 1290, true),
    ('00000000-0000-0000-0000-0000000000a5', 'STICKER-PACK', 'Sticker pack', 390, true),
    ('00000000-0000-0000-0000-0000000000a6', 'RETRO-POSTER', 'Retro poster', 890, false)
ON CONFLICT (sku) DO NOTHING;

INSERT INTO orders (id, user_id, status, note, total_cents)
SELECT '00000000-0000-0000-0000-0000000000b1', 1, 'pending', 'leave at the door', 1680
WHERE NOT EXISTS (SELECT 1 FROM orders WHERE id = '00000000-0000-0000-0000-0000000000b1');

INSERT INTO order_lines (order_id, item_id, qty, price_cents)
SELECT '00000000-0000-0000-0000-0000000000b1', '00000000-0000-0000-0000-0000000000a1', 1, 990
WHERE NOT EXISTS (SELECT 1 FROM order_lines WHERE order_id = '00000000-0000-0000-0000-0000000000b1');

INSERT INTO order_lines (order_id, item_id, qty, price_cents)
SELECT '00000000-0000-0000-0000-0000000000b1', '00000000-0000-0000-0000-0000000000a5', 1, 390
WHERE NOT EXISTS (
    SELECT 1 FROM order_lines
    WHERE order_id = '00000000-0000-0000-0000-0000000000b1'
      AND item_id = '00000000-0000-0000-0000-0000000000a5'
);

INSERT INTO orders (id, user_id, status, note, total_cents)
SELECT '00000000-0000-0000-0000-0000000000b2', 2, 'paid', '', 1290
WHERE NOT EXISTS (SELECT 1 FROM orders WHERE id = '00000000-0000-0000-0000-0000000000b2');

INSERT INTO order_lines (order_id, item_id, qty, price_cents)
SELECT '00000000-0000-0000-0000-0000000000b2', '00000000-0000-0000-0000-0000000000a4', 1, 1290
WHERE NOT EXISTS (SELECT 1 FROM order_lines WHERE order_id = '00000000-0000-0000-0000-0000000000b2');
