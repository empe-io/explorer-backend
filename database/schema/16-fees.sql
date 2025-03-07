CREATE TABLE fees
(
    height             BIGINT PRIMARY KEY,
    fee_value          BIGINT NOT NULL,
    stable_fee_value   BIGINT NOT NULL
);
CREATE INDEX fees_height_index ON fees (height);