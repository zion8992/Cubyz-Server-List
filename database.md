# Setup Database

This file will guide you through how to setup a MySQL database for ironite.
If you have Docker installed you can easily create and run a database with the scripts in `scripts/`.

If you don't, install MySQL or download it from [https://www.mysql.com/downloads/](mysql.com/downloads).

### Database Tables

```sql
create database ironite;


/** USERS **/

DROP TABLE IF EXISTS users;

CREATE TABLE users (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  username VARCHAR(100) NOT NULL,
  email VARCHAR(255) NOT NULL UNIQUE,
  password VARCHAR(255) NOT NULL,
  date_created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  session_token VARCHAR(255) NULL,
  session_token_expires TIMESTAMP NULL
  profile_picture_url VARCHAR(500) NULL,
  description TEXT NULL,
  pubkey TEXT NULL,
  pronouns VARCHAR(50) NULL;
);

/** SERVERS **/

DROP TABLE IF EXISTS servers;

CREATE TABLE servers (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  description TEXT,
  xml_feed_link TEXT,
  playercount BIGINT UNSIGNED NOT NULL DEFAULT 0,
  owner_id BIGINT UNSIGNED NOT NULL,
  gamemodes VARCHAR(255),
  version VARCHAR(255),
  languages VARCHAR(255),
  requires_mods BOOLEAN NOT NULL DEFAULT FALSE,
  website_url TEXT,
  chat_url TEXT,
  ip VARCHAR(255) NOT NULL DEFAULT '',
  last_spark TIMESTAMP NULL,
  status BOOLEAN NOT NULL DEFAULT FALSE;
  FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE
);


/** API TOKENS **/
CREATE TABLE api_tokens (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT UNSIGNED NOT NULL,
    server_id BIGINT UNSIGNED NOT NULL,
    date_created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expiry TIMESTAMP NOT NULL,
    type VARCHAR(50) NOT NULL,
    token_hash VARCHAR(255) NOT NULL,

    CONSTRAINT fk_api_tokens_owner
        FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_api_tokens_server
        FOREIGN KEY (server_id) REFERENCES servers(id) ON DELETE CASCADE,
    CONSTRAINT chk_expiry_after_created
        CHECK (expiry > date_created)
);

CREATE INDEX idx_api_tokens_owner_id ON api_tokens(owner_id);
CREATE INDEX idx_api_tokens_server_id ON api_tokens(server_id);
CREATE INDEX idx_api_tokens_expiry ON api_tokens(expiry);
CREATE INDEX idx_api_tokens_token_hash ON api_tokens(token_hash);
```

## Create test token

```sql
INSERT INTO api_tokens (
    id,
    owner_id,
    server_id,
    date_created,
    expiry,
    type,
    token_hash
) VALUES (
    1,
    1,
    3,
    CURRENT_TIMESTAMP,
    DATE_ADD(CURRENT_TIMESTAMP, INTERVAL 90 DAY),
    'spark',
    'snails'
);
```