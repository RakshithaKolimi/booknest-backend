-- Create Enum for user role --
CREATE TYPE USER_ROLE AS ENUM ('USER', 'ADMIN');

CREATE TYPE PAYMENT_STATUS AS ENUM ('PENDING', 'PAID', 'REFUND_INITIATED', 'REFUNDED', 'FAILED');

CREATE TYPE PAYMENT_METHOD AS ENUM (
  'COD',
  'CREDIT_CARD',
  'DEBIT_CARD',
  'NET_BANKING',
  'UPI'
);

CREATE TYPE ORDER_STATUS AS ENUM ('PENDING', 'FAILED', 'CANCELLED', 'COMPLETED');

-- Create table for users --
CREATE TABLE IF NOT EXISTS users (
  id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
  first_name VARCHAR(255) NOT NULL,
  last_name VARCHAR(255) NOT NULL,
  email VARCHAR(255) NOT NULL UNIQUE,
  mobile VARCHAR(255) NOT NULL UNIQUE,
  password VARCHAR(255) NOT NULL,
  last_login TIMESTAMP DEFAULT NULL,
  role USER_ROLE NOT NULL DEFAULT 'USER',
  is_active BOOLEAN NOT NULL DEFAULT false,
  email_verified BOOLEAN NOT NULL DEFAULT false,
  mobile_verified BOOLEAN NOT NULL DEFAULT false,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW(),
  deleted_at TIMESTAMP DEFAULT NULL
);

-- Create table for publishers --
CREATE TABLE IF NOT EXISTS publishers (
  id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
  legal_name VARCHAR(255) NOT NULL,
  trading_name VARCHAR(255) NOT NULL,
  email VARCHAR(255) NOT NULL,
  mobile VARCHAR(255) NOT NULL,
  address TEXT NOT NULL DEFAULT '',
  city VARCHAR(255) NOT NULL,
  state VARCHAR(255) NOT NULL,
  country VARCHAR(255) NOT NULL,
  zipcode VARCHAR(255) NOT NULL,
  is_active BOOLEAN NOT NULL DEFAULT false,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW(),
  deleted_at TIMESTAMP DEFAULT NULL
);

-- Create table for books --
CREATE TABLE IF NOT EXISTS books (
  id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
  name VARCHAR(255) NOT NULL,
  author_name VARCHAR(255) NOT NULL,
  available_stock INT NOT NULL DEFAULT 0 CHECK (available_stock >= 0),
  image_url VARCHAR(255),
  is_active BOOLEAN DEFAULT false,
  description VARCHAR(255) DEFAULT '',
  isbn VARCHAR(255) UNIQUE,
  price NUMERIC(10, 2) DEFAULT 0.00,
  discount_percentage NUMERIC(10, 2) CHECK (
    discount_percentage >= 0
    AND discount_percentage <= 100
  ),
  publisher_id UUID NOT NULL,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW(),
  deleted_at TIMESTAMP DEFAULT NULL,
  -- Foreign key --
  FOREIGN KEY (publisher_id) REFERENCES publishers(id) ON DELETE RESTRICT
);

-- Create table for orders --
CREATE TABLE IF NOT EXISTS orders (
  id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
  order_number VARCHAR(255) NOT NULL UNIQUE,
  total_price NUMERIC(10, 2) DEFAULT 0.00,
  user_id UUID NOT NULL,
  payment_method PAYMENT_METHOD DEFAULT NULL,
  payment_status PAYMENT_STATUS DEFAULT NULL,
  status ORDER_STATUS DEFAULT 'PENDING',
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW(),
  deleted_at TIMESTAMP DEFAULT NULL,
  -- Foreign key --
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT
);

-- Create table for order_items --
CREATE TABLE IF NOT EXISTS order_items (
  order_id UUID NOT NULL,
  book_id UUID NOT NULL,
  purchase_count INT NOT NULL DEFAULT 1 CHECK (purchase_count > 0),
  purchase_price NUMERIC(10, 2) DEFAULT 0.00,
  total_price NUMERIC(10, 2) DEFAULT 0.00,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW(),
  deleted_at TIMESTAMP DEFAULT NULL,
  -- Primary keys --
  PRIMARY KEY (order_id, book_id),
  -- Foreign keys --
  FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE,
  FOREIGN KEY (book_id) REFERENCES books(id) ON DELETE CASCADE
);

-- Create table for cart --
CREATE TABLE IF NOT EXISTS carts (
  id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL,
  -- Foreign key --
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT
);

-- Create table for cart_items --
CREATE TABLE IF NOT EXISTS cart_items (
  cart_id UUID NOT NULL,
  count INT NOT NULL DEFAULT 1 CHECK (count > 0),
  book_id UUID NOT NULL,
  cart_price NUMERIC(10, 2) DEFAULT 0.00,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW(),
  deleted_at TIMESTAMP DEFAULT NULL,
  -- Primary keys --
  PRIMARY KEY (cart_id, book_id),
  -- Foreign keys --
  FOREIGN KEY (cart_id) REFERENCES carts(id) ON DELETE CASCADE,
  FOREIGN KEY (book_id) REFERENCES books(id) ON DELETE CASCADE
);

-- Create table for categories --
CREATE TABLE IF NOT EXISTS categories (
  id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
  name VARCHAR(255) NOT NULL,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW(),
  deleted_at TIMESTAMP DEFAULT NULL
);

-- Create table for book categories --
CREATE TABLE IF NOT EXISTS book_categories (
  book_id UUID NOT NULL,
  category_id UUID NOT NULL,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW(),
  deleted_at TIMESTAMP DEFAULT NULL,
  -- Primary keys --
  PRIMARY KEY (book_id, category_id),
  -- Foreign keys --
  FOREIGN KEY (book_id) REFERENCES books(id) ON DELETE CASCADE,
  FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE
);

-- Create Indices --
CREATE INDEX idx_books_publisher_id ON books(publisher_id);

CREATE INDEX idx_orders_user_id ON orders(user_id);

CREATE INDEX idx_cart_user_id ON carts(user_id);

CREATE INDEX idx_order_items_book_id ON order_items(book_id);

CREATE INDEX idx_cart_items_book_id ON cart_items(book_id);

-- Add cart constraints --
ALTER TABLE
  carts
ADD
  CONSTRAINT unique_user_cart UNIQUE (user_id);
