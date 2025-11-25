-- ========================================
-- PostgreSQL schema for YasaiRap (minimal)
-- whitelist_users only
-- ========================================

CREATE TABLE whitelist_users (
  id               BIGSERIAL PRIMARY KEY,
  discord_user_id  VARCHAR(64)  NOT NULL,
  vrc_user_id      VARCHAR(64)  NOT NULL,
  vrc_display_name VARCHAR(64)  NOT NULL,
  vrc_avatar_url   VARCHAR(512),
  note             VARCHAR(255) NOT NULL DEFAULT '',
  created_at       TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at       TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX uq_discord_user ON whitelist_users (discord_user_id);
CREATE UNIQUE INDEX uq_vrc_user     ON whitelist_users (vrc_user_id);
