-- +migrate Up

CREATE TABLE IF NOT EXISTS blocks
(
    number   INT UNSIGNED PRIMARY KEY,
    hash     VARCHAR(66),
    parentHash VARCHAR(66),
    miner    VARCHAR(42),
    gasLimit BIGINT UNSIGNED,
    gasUsed  BIGINT UNSIGNED,
    blockTimestamp TIMESTAMP,
    transactionsCount INT
);

REPLACE INTO blocks(number) values(0);

CREATE TABLE IF NOT EXISTS transactions
(
    hash            VARCHAR(66) PRIMARY KEY,
    nonce           INT,
    blockNumber     INT UNSIGNED,
    transactionIndex INT UNSIGNED,
    fromAddress     VARCHAR(42),
    toAddress       VARCHAR(42),
    value           VARCHAR(80),
    gas             VARCHAR(80),
    gasPrice        VARCHAR(80),
    input           TEXT,
    contractAddress VARCHAR(42),
    type            ENUM('transfer', 'approve'),

    FOREIGN KEY (blockNumber) REFERENCES blocks(number)
);

CREATE TABLE IF NOT EXISTS exchanges
(
    address VARCHAR(42) PRIMARY KEY,
    name VARCHAR(100)
);

CREATE TABLE IF NOT EXISTS exchangeTransfers
(
    txDeposit VARCHAR(66) PRIMARY KEY,
    txExchange VARCHAR(66),

    FOREIGN KEY (txDeposit) REFERENCES transactions(hash),
    FOREIGN KEY (txExchange) REFERENCES transactions(hash)
);

CREATE TABLE IF NOT EXISTS clusters
(
    id INT AUTO_INCREMENT PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS accounts
(
    address VARCHAR(42) PRIMARY KEY,
    accountType ENUM('eoa', 'miner', 'exchange', 'deposit', 'scammer'),
    comment TEXT,
    cluster INT,

    FOREIGN KEY (cluster) REFERENCES clusters(id)
);

SET GLOBAL sql_mode=(SELECT REPLACE(@@sql_mode,'ONLY_FULL_GROUP_BY',''));

-- +migrate Down
DROP TABLE IF EXISTS exchangeTransfers;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS blocks;

DROP TABLE IF EXISTS exchanges;
DROP TABLE IF EXISTS accounts;
DROP TABLE IF EXISTS clusters;
