PRAGMA journal_mode=WAL;
PRAGMA synchronous=NORMAL;
PRAGMA foreign_keys=ON;
PRAGMA cache_size=-64000;

CREATE TABLE IF NOT EXISTS user (
    id                  INTEGER         NOT NULL PRIMARY KEY,                       -- Account ID
    created             TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,         -- Created At
    updated             TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,         -- Updated At

    -- Account
    email_address       TEXT            NOT NULL UNIQUE,                            -- User Email Address
    email_verified      BOOLEAN         NOT NULL DEFAULT 0,                         -- Email Verified?
    ip_address          TEXT            NOT NULL DEFAULT '',                        -- New Login IP Address
    mfa_enabled         BOOLEAN         NOT NULL DEFAULT 0,                         -- MFA Enabled?
    mfa_secret          TEXT,                                                       -- MFA Secret Key
    mfa_codes           TEXT            NOT NULL DEFAULT '',                        -- [ARRAY] MFA Recovery Codes
    mfa_codes_used      INT             NOT NULL DEFAULT 0,                         -- MFA Exhausted Recovery Code Bitfield
    password_hash       TEXT,                                                       -- Active Password Hash
    password_history    TEXT            NOT NULL DEFAULT '',                        -- [ARRAY] Past Password Hashes
    token_verify        TEXT            UNIQUE,                                     -- Verify Email Token
    token_verify_eat    TIMESTAMP,                                                  -- Verify Email Token Expires At
    token_login         TEXT            UNIQUE,                                     -- New Login Token
    token_login_data    TEXT,                                                       -- New Login Token Arbitrary Data
    token_login_eat     TIMESTAMP,                                                  -- New Login Token Expires At
    token_reset         TEXT            UNIQUE,                                     -- Password Reset Token
    token_reset_eat     TIMESTAMP,                                                  -- Password Reset Token Expires At
    token_passcode      TEXT,                                                       -- Random Escalation Code
    token_passcode_eat  TIMESTAMP,                                                  -- Random Escalation Code Expires At

    -- Profile
    username            TEXT            NOT NULL UNIQUE,                            -- Username
    displayname         TEXT            NOT NULL,                                   -- Nickname
    subtitle            TEXT,                                                       -- Pronouns
    biography           TEXT,                                                       -- Biography
    avatar_hash         TEXT,                                                       -- Avatar Image Hash
    banner_hash         TEXT,                                                       -- Banner Image Hash
    accent_banner       INT,                                                        -- Banner Color
    accent_border       INT,                                                        -- Border Color
    accent_background   INT                                                         -- Background Color
);

CREATE TABLE IF NOT EXISTS user_session (
    id                  INTEGER         NOT NULL PRIMARY KEY,                       -- Session ID
    created             TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,         -- Created At
    updated             TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,         -- Updated At
    user_id             INTEGER         NOT NULL,                                   -- Relevant User ID
    token               TEXT            UNIQUE,                                     -- Session Token
    elevated_until      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,         -- Elevated Until UNIX Timestamp
    device_ip_address   TEXT            NOT NULL,                                   -- IP Address of Device
    device_user_agent   TEXT            NOT NULL,                                   -- User Agent of Device
    device_public_key   TEXT            NOT NULL,                                   -- Device Public Key
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
