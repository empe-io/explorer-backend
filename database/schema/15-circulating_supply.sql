CREATE TABLE circulating_supply
(
    height BIGINT PRIMARY KEY,
    value  BIGINT NOT NULL
);
CREATE INDEX circulating_supply_height_index ON circulating_supply (height);