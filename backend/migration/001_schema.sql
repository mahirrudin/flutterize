CREATE TABLE IF NOT EXISTS users (
    id                  BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    email               VARCHAR(255) NOT NULL UNIQUE,
    phone               VARCHAR(20)  NOT NULL UNIQUE,
    fullname            VARCHAR(255) NOT NULL,
    birthdate           DATE         NOT NULL,
    password            VARCHAR(255) NOT NULL,
    points_balance      BIGINT       NOT NULL DEFAULT 0,
    reset_token         VARCHAR(255) DEFAULT NULL,
    reset_token_expiry  DATETIME     DEFAULT NULL,
    created_at          TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP    DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS point_transactions (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    sender_id       BIGINT UNSIGNED NOT NULL,
    receiver_id     BIGINT UNSIGNED NOT NULL,
    amount          BIGINT NOT NULL,
    note            VARCHAR(255) DEFAULT NULL,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (sender_id)   REFERENCES users(id),
    FOREIGN KEY (receiver_id) REFERENCES users(id)
);
