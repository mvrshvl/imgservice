-- +migrate Up

CREATE TABLE IF NOT EXISTS blocks
(
    number   INT UNSIGNED PRIMARY KEY,
    hash     BINARY(64),
    parentHash     BINARY(64),
    nonce    INT UNSIGNED,
    miner    BINARY(20),
    gasLimit BIGINT,
    gasUsed  BIGINT,
    blockTimestamp TIMESTAMP,
    transactionsCount INT
);

INSERT INTO blocks(number) values(0);

CREATE TABLE IF NOT EXISTS transactions
(
    hash            BINARY(32) PRIMARY KEY,
    nonce           INT,
    blockNumber     INT UNSIGNED,
    fromAddress     BINARY(20),
    toAddress       BINARY(20),
    value           INTEGER,
    gas             INTEGER,
    gasPrice        INTEGER,
    contractAddress BINARY(20),
    event           ENUM('transfer', 'approve'),

    FOREIGN KEY (blockNumber) REFERENCES blocks(number)
);

CREATE TABLE IF NOT EXISTS exchanges
(
    address BINARY(20) PRIMARY KEY,
    name VARCHAR(100)
);

CREATE TABLE IF NOT EXISTS clusters
(
    id INT AUTO_INCREMENT PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS accounts
(
    address BINARY(20) PRIMARY KEY,
    accountType ENUM('eoa', 'miner'),
    cluster INT,

    FOREIGN KEY (cluster) REFERENCES clusters(id)
);

-- +migrate Down
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS blocks;

DROP TABLE IF EXISTS exchanges;
DROP TABLE IF EXISTS accounts;
DROP TABLE IF EXISTS clusters;
