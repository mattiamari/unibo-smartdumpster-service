CREATE SCHEMA IF NOT EXISTS smartdumpster;

CREATE TYPE dumptype as ENUM ('paper', 'plastic', 'unsorted');

CREATE TABLE smartdumpster.dumpster (
    id uuid NOT NULL,
    name varchar,
    available boolean DEFAULT false,
    weight_limit integer NOT NULL,
    PRIMARY KEY (id)
);

CREATE TABLE smartdumpster.dump (
    id_user uuid NOT NULL,
    id_dumpster uuid NOT NULL,
    created_at timestamptz DEFAULT CURRENT_TIMESTAMP NOT NULL,
    type dumptype NOT NULL,
    PRIMARY KEY (id_dumpster, created_at),
    FOREIGN KEY (id_dumpster) REFERENCES dumpster(id)
);

CREATE TABLE smartdumpster.weight (
    id_dumpster uuid NOT NULL,
    weight integer NOT NULL,
    created_at timestamptz DEFAULT CURRENT_TIMESTAMP NOT NULL,
    PRIMARY KEY (id_dumpster, created_at),
    FOREIGN KEY (id_dumpster) REFERENCES dumpster(id)
);

INSERT INTO smartdumpster.dumpster VALUES ('c37246d2-3088-45eb-af40-87dfdd6bf314', 'Test Dumpster', true, 100);
INSERT INTO smartdumpster.dumpster VALUES ('4dcd70bb-4d32-464f-bda0-fd28eb5fe403', 'Test Dumpster 2', true, 200);