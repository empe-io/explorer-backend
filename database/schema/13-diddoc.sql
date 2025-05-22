CREATE TABLE did_document
(
    did    TEXT   NOT NULL,
    height BIGINT NOT NULL,
    json   TEXT   NOT NULL
);
CREATE INDEX did_document_id_index ON did_document (did);
CREATE INDEX did_document_height_index ON did_document (height);
