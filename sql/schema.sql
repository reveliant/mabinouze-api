DROP TABLE IF EXISTS drinks;
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS rounds;

CREATE TABLE rounds (
  round_id uuid PRIMARY KEY,
  name varchar(255) UNIQUE NOT NULL,
  description varchar(255) NOT NULL,
  time timestamp with time zone NOT NULL,
  expires timestamp with time zone NOT NULL,
  password char(76) NOT NULL,
  access_token char(76) NULL,
  locked boolean NOT NULL DEFAULT FALSE
);

CREATE TABLE orders (
  order_id uuid PRIMARY KEY,
  round_id uuid NOT NULL,
  name varchar(255) NOT NULL,
  password char(76) NOT NULL,
  FOREIGN KEY (round_id) REFERENCES rounds (round_id),
  UNIQUE(round_id, name)
);

CREATE TABLE drinks (
  drink_id uuid PRIMARY KEY,
  order_id uuid NOT NULL,
  name varchar(255) NOT NULL,
  quantity smallint NOT NULL,
  FOREIGN KEY (order_id) REFERENCES orders (order_id),
  UNIQUE(order_id, name)
);
