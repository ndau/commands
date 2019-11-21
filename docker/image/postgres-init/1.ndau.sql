-- define some users
CREATE USER guest;
CREATE USER node WITH PASSWORD NULL; -- can't log in until password is set

-- define the databse
CREATE DATABASE ndau
    WITH
        OWNER = node
;
\connect ndau

-- define some tables
CREATE TABLE Blocks (
    height INTEGER PRIMARY KEY,
    block_time TIMESTAMP
        WITHOUT TIME ZONE -- all our times are UTC
        NOT NULL
);

CREATE TABLE Transactions (
    id SERIAL PRIMARY KEY,
    tx_name TEXT -- in postgres, this is as efficient as varchar(n)
        NOT NULL
        CHECK (tx_name <> ''),
    block INTEGER REFERENCES Blocks(height)
        NOT NULL,
    sequence SMALLINT NOT NULL DEFAULT 0,
    data JSONB
        NOT NULL
        CHECK (char_length(data::text) > 0),
    UNIQUE (block, sequence)
);

CREATE TABLE Accounts (
    id SERIAL PRIMARY KEY,
    address TEXT
        NOT NULL
        check(address <> ''),
    tx INTEGER REFERENCES Transactions(id)
        NOT NULL,
    data JSONB
        NOT NULL
        CHECK (char_length(data::text) > 0),
    UNIQUE (address, tx)
);

CREATE TABLE SystemVariables (
    id SERIAL PRIMARY KEY,
    tx INTEGER REFERENCES Transactions(id)
        NULL -- genesis state when null
        UNIQUE,
    key TEXT
        NOT NULL
        CHECK (key <> ''),
    value BYTEA -- like TEXT, but handles raw binary data
        NOT NULL
        CHECK (value <> '')
);

-- define some useful indices
CREATE INDEX ON Blocks(block_time);
CREATE INDEX ON Transactions(tx_name);
CREATE INDEX ON Accounts(address);
CREATE INDEX ON SystemVariables(key);

-- now set up the permissions
GRANT SELECT, INSERT ON ALL TABLES IN SCHEMA public TO node;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO guest;
