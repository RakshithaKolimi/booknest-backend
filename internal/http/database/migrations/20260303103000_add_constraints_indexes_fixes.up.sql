-- Ensure verification tokens are tied to a valid user
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'fk_verification_tokens_user_id'
  ) THEN
    ALTER TABLE verification_tokens
    ADD CONSTRAINT fk_verification_tokens_user_id
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
  END IF;
END $$;

-- Indexes for common query predicates and sorting
CREATE INDEX IF NOT EXISTS idx_books_active_created_at
ON books(is_active, created_at DESC)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_books_price_not_deleted
ON books(price)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_book_categories_category_id_not_deleted
ON book_categories(category_id)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_orders_user_id_created_at
ON orders(user_id, created_at DESC)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_cart_items_cart_id_not_deleted
ON cart_items(cart_id)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_verification_tokens_lookup_active
ON verification_tokens(type, token_hash, expires_at)
WHERE is_used = FALSE;

CREATE INDEX IF NOT EXISTS idx_authors_lower_name
ON authors(LOWER(name));

CREATE INDEX IF NOT EXISTS idx_categories_lower_name
ON categories(LOWER(name));
