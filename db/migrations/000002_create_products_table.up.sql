CREATE TABLE IF NOT EXISTS products(
    id serial PRIMARY KEY,
    seller_id char(36) NOT NULL,
    name varchar(30) NOT NULL,
    product_picture varchar(255) NOT NULL,
    quantity int NOT NULL,
    price decimal(10,2) NOT NULL,
    weight int NOT NULL,
    size varchar(4) NOT NULL,
    status varchar(9) DEFAULT 'Ready',
    description text NOT NULL,
    created_at timestamp NOT NULL,
    updated_at timestamp NOT NULL,
    FOREIGN KEY (seller_id) REFERENCES users(id) ON DELETE CASCADE
)