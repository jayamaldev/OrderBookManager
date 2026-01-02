CREATE TYPE userStatus AS ENUM ('Active', 'Inactive');

CREATE TABLE users (
  userId SERIAL PRIMARY KEY,
  firstName varchar(50) NOT NULL,
  lastName varchar(50) NOT NULL,
  email varchar NOT NULL,
  phone varchar,
  age int,
  user_status userStatus DEFAULT 'Active'
);