# Setup Database
```
create database ironite;
```


```sql
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
  FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE
);
```