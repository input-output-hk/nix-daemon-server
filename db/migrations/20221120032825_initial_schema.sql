-- migrate:up

CREATE TABLE valid_paths (
    id INTEGER PRIMARY KEY NOT NULL,
    path TEXT UNIQUE NOT NULL,
    hash TEXT NOT NULL,
    registration_time TIMESTAMP WITH TIME ZONE,
    deriver TEXT,
    nar_size INTEGER,
    ultimate BOOLEAN,
    sigs TEXT[],
    ca TEXT -- if not null, an assertion that the path is content-addressed; see ValidPathInfo
);

create table refs (
    referrer  INTEGER NOT NULL,
    reference INTEGER NOT NULL,
    PRIMARY KEY (referrer, reference),
    FOREIGN KEY (referrer) REFERENCES valid_paths(id) ON DELETE CASCADE,
    FOREIGN KEY (reference) REFERENCES valid_paths(id) ON DELETE RESTRICT
);

CREATE INDEX index_referrer ON refs(referrer);
CREATE INDEX index_reference ON refs(reference);

CREATE TABLE derivation_outputs (
    drv  INTEGER NOT NULL,
    id   TEXT NOT NULL, -- symbolic output id, usually "out"
    path TEXT NOT NULL,
    PRIMARY KEY (drv, id),
    FOREIGN KEY (drv) REFERENCES valid_paths(id) ON DELETE CASCADE
);

CREATE INDEX index_derivation_outputs ON derivation_outputs(path);

-- migrate:down

DROP TABLE derivation_outputs;
DROP TABLE refs;
DROP TABLE valid_paths;
