-- Create "whitelist_users" table
CREATE TABLE "public"."whitelist_users" (
  "id" bigserial NOT NULL,
  "discord_user_id" character varying(64) NOT NULL,
  "vrc_user_id" character varying(64) NOT NULL,
  "vrc_display_name" character varying(64) NOT NULL,
  "vrc_avatar_url" character varying(512) NULL,
  "note" character varying(255) NOT NULL DEFAULT '',
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id")
);
-- Create index "uq_discord_user" to table: "whitelist_users"
CREATE UNIQUE INDEX "uq_discord_user" ON "public"."whitelist_users" ("discord_user_id");
-- Create index "uq_vrc_user" to table: "whitelist_users"
CREATE UNIQUE INDEX "uq_vrc_user" ON "public"."whitelist_users" ("vrc_user_id");
