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
        NOT NULL,
    hash TEXT
        NOT NULL
        CHECK (hash <> '')
        UNIQUE
);

CREATE TABLE Transactions (
    id SERIAL PRIMARY KEY,
    name TEXT -- in postgres, this is as efficient as varchar(n)
        NOT NULL
        CHECK (name <> ''),
    hash TEXT -- in postgres, this is as efficient as varchar(n)
        NOT NULL
        CHECK (hash <> '')
        UNIQUE,
    height INTEGER REFERENCES Blocks(height)
        NOT NULL,
    sequence SMALLINT NOT NULL DEFAULT 0,
    data JSONB
        NOT NULL
        CHECK (char_length(data::text) > 0),
    fee BIGINT NOT NULL,
    sib BIGINT NOT NULL,
    UNIQUE (height, sequence)
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
    -- height isn't a reference because we expect to capture some genesis state.
    -- genesis data has height 0, which is never a valid block number.
    height INTEGER
        NOT NULL, -- genesis state when 0
    key TEXT
        NOT NULL
        CHECK (key <> ''),
    value BYTEA -- like TEXT, but handles raw binary data
        NOT NULL
        CHECK (value <> '')
);

CREATE TABLE MarketPrices (
    tx INTEGER PRIMARY KEY REFERENCES Transactions(id),
    price BIGINT NOT NULL
);

CREATE TABLE TargetPrices (
    tx INTEGER PRIMARY KEY REFERENCES Transactions(id),
    price BIGINT NOT NULL
);

-- define some useful indices
CREATE INDEX ON Accounts(address);
CREATE INDEX ON Blocks(block_time);
CREATE INDEX ON Blocks(hash);
CREATE INDEX ON SystemVariables(height);
CREATE INDEX ON SystemVariables(key);
CREATE INDEX ON Transactions(hash);
CREATE INDEX ON Transactions(name);

-- now set up the permissions
GRANT SELECT, INSERT ON ALL TABLES IN SCHEMA public TO node;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO node;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO guest;
