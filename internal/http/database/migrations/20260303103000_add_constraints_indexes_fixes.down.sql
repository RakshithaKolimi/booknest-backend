DROP INDEX IF EXISTS idx_categories_lower_name;
DROP INDEX IF EXISTS idx_authors_lower_name;
DROP INDEX IF EXISTS idx_verification_tokens_lookup_active;
DROP INDEX IF EXISTS idx_cart_items_cart_id_not_deleted;
DROP INDEX IF EXISTS idx_orders_user_id_created_at;
DROP INDEX IF EXISTS idx_book_categories_category_id_not_deleted;
DROP INDEX IF EXISTS idx_books_price_not_deleted;
DROP INDEX IF EXISTS idx_books_active_created_at;

ALTER TABLE verification_tokens
DROP CONSTRAINT IF EXISTS fk_verification_tokens_user_id;

ALTER TABLE books
DROP CONSTRAINT IF EXISTS uq_books_isbn;

ALTER TABLE users
DROP CONSTRAINT IF EXISTS uq_users_email;
