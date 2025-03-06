CREATE TABLE circulating_supply
(
    one_row_id bool PRIMARY KEY DEFAULT TRUE,
    value      BIGINT NOT NULL,
    height     BIGINT  NOT NULL,
    CONSTRAINT one_row_uni CHECK (one_row_id)
);
CREATE INDEX circulating_supply_height_index ON circulating_supply (height);