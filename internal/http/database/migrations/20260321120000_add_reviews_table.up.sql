CREATE TABLE IF NOT EXISTS reviews (
  id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
  book_id UUID NOT NULL,
  user_id UUID NOT NULL,
  rating INT NOT NULL CHECK (rating >= 1 AND rating <= 5),
  comment TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW(),
  deleted_at TIMESTAMP DEFAULT NULL,
  CONSTRAINT fk_reviews_book_id FOREIGN KEY (book_id) REFERENCES books(id) ON DELETE CASCADE,
  CONSTRAINT fk_reviews_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  CONSTRAINT ux_reviews_book_user UNIQUE (book_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_reviews_book_id_created_at
ON reviews(book_id, created_at DESC)
WHERE deleted_at IS NULL;
