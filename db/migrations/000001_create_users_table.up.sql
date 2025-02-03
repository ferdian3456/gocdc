CREATE TABLE users(
    id char(36) PRIMARY KEY,
    name varchar(20) UNIQUE NOT NULL,
    email varchar(254) UNIQUE NOT NULL,
    password varchar(60) NOT NULL,
    address varchar(30) NOT NULL,
    phonenumber varchar(12) NOT NULL,
    created_at timestamp NOT NULL,
    updated_at timestamp NOT NULL
)