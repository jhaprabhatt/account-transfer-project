CREATE TABLE IF NOT EXISTS accounts
(
    account_id BIGINT PRIMARY KEY,
    balance    NUMERIC(20, 5) NOT NULL DEFAULT 0,
    CONSTRAINT check_balance_positive CHECK (balance >= 0)
);

CREATE TABLE IF NOT EXISTS transfers
(
    transfer_id              SERIAL PRIMARY KEY,
    correlation_id           BIGINT         NOT NULL,
    status                   INT                      DEFAULT 1, -- 1: PENDING, 2: COMPLETED, 3: FAILED
    source_account_id        BIGINT         NOT NULL,
    destination_account_id   BIGINT         NOT NULL,
    amount                   NUMERIC(20, 5) NOT NULL,
    source_prev_balance      NUMERIC(20, 5) NOT NULL,
    source_post_balance      NUMERIC(20, 5)           DEFAULT 0,
    destination_prev_balance NUMERIC(20, 5) NOT NULL,
    destination_post_balance NUMERIC(20, 5)           DEFAULT 0,
    created_at               TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_source FOREIGN KEY (source_account_id) REFERENCES accounts (account_id),
    CONSTRAINT fk_dest FOREIGN KEY (destination_account_id) REFERENCES accounts (account_id)
);

CREATE INDEX IF NOT EXISTS idx_transfers_source ON transfers (source_account_id);
CREATE INDEX IF NOT EXISTS idx_transfers_dest ON transfers (destination_account_id);