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
    accountType INT UNSIGNED,
    comment TEXT,
    cluster INT,

    FOREIGN KEY (cluster) REFERENCES clusters(id)
);

CREATE TABLE IF NOT EXIST accountType
(
--     accountType ENUM('eoa', 'miner', 'exchange', 'deposit', 'scammer')
    idx INT UNSIGNED PRIMARY KEY,
    name VARCHAR(50),
    descriprion VARCHAR(200),
    risk INT UNSIGNED,
)

INSERT INTO accountType(idx, name, risk)
    VALUES  (0, 'eoa', 'Untyped account', 0),
            (1, 'miner', 'Coins mined by miners and not yet forwarded', 0),
            (2, 'payment management', 'Coins associated with payment services', 5),
            (3, 'wallet', 'Coins stored in verified wallets', 10),
            (4, 'exchange', 'exchanges that require KYC/AML identification for all deposit or withdrawal', 15),
            (5, 'p2p exchange low risk', 'P2P exchanges that require KYC/AML identification for all deposits and withdrawals', 20),
            (6, 'marketplace', 'Coins that were used to pay for legal activities', 24),
            (7, 'p2p exchange high risk', 'P2P exchanges that allow the withdrawal of more than $1000 in crypto daily without KYC/ AML', 30),
            (8, 'exchange risk high', 'exchanges that allow the withdrawal of more than $2000 in crypto daily without KYC/AML. (For fiat withdrawals, KYC/AML is still required)', 40),
            (9, 'exchange risk moderate', 'exchanges that allow the withdrawal of up to $2000 in crypto daily without KYC/AML. (For fiat withdrawals, KYC/AML is still required)', 50),
            (10, 'ATM', 'Coins obtained from a cryptocurrency ATM', 60),
            (11, 'exchange risk very high', 'exchanges that don’t use verification procedures, or have requirements for certain countries only', 70),
            (12, 'mixer', 'Coins that were passed through a mixer to make tracking difficult or impossible. Mixers are mainly used for money laundering', 75),
            (13, 'gambling', 'Coins associated with unlicensed online gaming', 78),
            (14, 'scam', 'coins that were obtained by deception', 81),
            (15, 'stolen', 'Stolen coins', 84),
            (16, 'exchange fraudulent', 'exchanges involved in exit scams, illegal behavior, or who have had funds seized by the government', 87),
            (17, 'ransom', 'Coins obtained through extortion or blackmail', 90),
            (18, 'illegal Service', 'Coins associated with illegal activities', 93),
            (19, 'dark market', 'coins that were used for shopping on the darknet', 96),
            (20, 'dark Service', 'coins related to child abuse, terrorist financing or drug trafficking', 100);

SET GLOBAL sql_mode=(SELECT REPLACE(@@sql_mode,'ONLY_FULL_GROUP_BY',''));

-- +migrate Down
DROP TABLE IF EXISTS exchangeTransfers;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS blocks;

DROP TABLE IF EXISTS exchanges;
DROP TABLE IF EXISTS accounts;
DROP TABLE IF EXISTS clusters;
