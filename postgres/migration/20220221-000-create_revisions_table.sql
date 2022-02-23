CREATE TABLE revisions
(
    id BIGSERIAL PRIMARY KEY,
    rev INT NOT NULL,
    version VARCHAR(255) NOT NULL,
    contract_id BIGINT NOT NULL REFERENCES contracts(id) ON DELETE CASCADE,
    notes TEXT NOT NULL DEFAULT '',
    code TEXT NOT NULL DEFAULT '',
    compiled_code BYTEA NOT NULL,
    max_fuel INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()

);

CREATE UNIQUE INDEX UQ_CONTRACT_REVISION on revisions(rev, contract_id);