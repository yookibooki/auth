CREATE TABLE users (
  id        SERIAL PRIMARY KEY,
  email     VARCHAR(320) NOT NULL UNIQUE,
  pwd_hash  CHAR(60) NOT NULL -- bcrypt
);

CREATE TABLE auth_codes (
  id            SERIAL PRIMARY KEY,
  code_hash     VARCHAR(60) NOT NULL UNIQUE, -- bcrypt
  user_id       INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  client_id     VARCHAR(64) NOT NULL,
  redirect_uri  VARCHAR(2048) NOT NULL,
  state         VARCHAR(512) NOT NULL,
  expires_at    TIMESTAMPTZ NOT NULL,
  used_at       TIMESTAMPTZ
);

CREATE INDEX auth_codes_exp_idx ON auth_codes(expires_at);
CREATE INDEX auth_codes_uid_idx ON auth_codes(user_id);

CREATE TABLE pwd_reset_tokens (
  id          SERIAL PRIMARY KEY,
  token_hash  VARCHAR(60) NOT NULL UNIQUE, -- bcrypt
  user_id     INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  expires_at  TIMESTAMPTZ NOT NULL,
  used_at     TIMESTAMPTZ
);

CREATE INDEX pwd_reset_exp_idx ON pwd_reset_tokens(expires_at);
CREATE INDEX pwd_reset_uid_idx ON pwd_reset_tokens(user_id);
