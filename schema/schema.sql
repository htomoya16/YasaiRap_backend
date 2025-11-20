-- Create ENUM types
CREATE TYPE ruleset_type AS ENUM ('league', 'single_elim', 'double_elim');
CREATE TYPE season_status_type AS ENUM ('draft', 'open', 'running', 'closed');
CREATE TYPE entry_kind_type AS ENUM ('solo', 'team');
CREATE TYPE asset_kind_type AS ENUM ('standings', 'schedule', 'profile');
CREATE TYPE bracket_type AS ENUM ('main', 'losers');
CREATE TYPE match_status_type AS ENUM ('pending', 'scheduled', 'completed', 'canceled');
CREATE TYPE result_method_type AS ENUM ('normal', 'ff', 'dq');
CREATE TYPE cypher_status_type AS ENUM ('playing', 'stopped');
CREATE TYPE tournament_status_type AS ENUM ('recruiting', 'ongoing', 'finished', 'canceled');
CREATE TYPE tournament_match_status_type AS ENUM ('pending', 'ongoing', 'voting', 'finished', 'canceled');
CREATE TYPE vrc_registration_status_type AS ENUM ('pending', 'approved', 'rejected');
CREATE TYPE vrc_event_type AS ENUM ('submitted', 'approved', 'rejected', 'revoked');

-- Create "whitelist_users" table
CREATE TABLE "whitelist_users" (
  "id" BIGSERIAL NOT NULL,
  "discord_user_id" VARCHAR(64) NOT NULL,
  "vrc_user_id" VARCHAR(64) NOT NULL,
  "vrc_display_name" VARCHAR(64) NOT NULL,
  "vrc_avatar_url" VARCHAR(512) NULL,
  "note" VARCHAR(255) NOT NULL DEFAULT '',
  "created_at" TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "uq_discord_user" UNIQUE ("discord_user_id"),
  CONSTRAINT "uq_vrc_user" UNIQUE ("vrc_user_id")
);

-- Create "users" table
CREATE TABLE "users" (
  "id" BIGSERIAL NOT NULL,
  "discord_id" VARCHAR(32) NOT NULL,
  "name" VARCHAR(191) NULL,
  "avatar_url" VARCHAR(255) NULL,
  "created_at" TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "uq_users_discord" UNIQUE ("discord_id")
);

-- Create "beats" table
CREATE TABLE "beats" (
  "id" BIGSERIAL NOT NULL,
  "title" VARCHAR(255) NOT NULL,
  "youtube_url" VARCHAR(500) NOT NULL,
  "video_id" VARCHAR(32) NOT NULL,
  "is_enabled" BOOLEAN NOT NULL DEFAULT true,
  "created_by" BIGINT NOT NULL,
  "created_at" TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "uq_beats_video_id" UNIQUE ("video_id"),
  CONSTRAINT "fk_beats_created_by" FOREIGN KEY ("created_by") REFERENCES "users" ("id") ON UPDATE RESTRICT ON DELETE RESTRICT
);
CREATE INDEX "idx_beats_created_by" ON "beats" ("created_by");
CREATE INDEX "idx_beats_is_enabled" ON "beats" ("is_enabled");

-- Create "servers" table
CREATE TABLE "servers" (
  "id" BIGSERIAL NOT NULL,
  "discord_guild_id" VARCHAR(32) NOT NULL,
  "name" VARCHAR(191) NULL,
  "icon_url" VARCHAR(255) NULL,
  "owner_discord_id" VARCHAR(32) NULL,
  "created_at" TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "uq_servers_discord" UNIQUE ("discord_guild_id")
);

-- Create "cypher_sessions" table
CREATE TABLE "cypher_sessions" (
  "id" BIGSERIAL NOT NULL,
  "server_id" BIGINT NOT NULL,
  "channel_id" VARCHAR(32) NOT NULL,
  "message_id" VARCHAR(32) NULL,
  "beat_id" BIGINT NOT NULL,
  "started_by" BIGINT NOT NULL,
  "status" cypher_status_type NOT NULL DEFAULT 'playing',
  "started_at" TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "ended_at" TIMESTAMP(6) NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_cypher_beat" FOREIGN KEY ("beat_id") REFERENCES "beats" ("id") ON UPDATE RESTRICT ON DELETE RESTRICT,
  CONSTRAINT "fk_cypher_server" FOREIGN KEY ("server_id") REFERENCES "servers" ("id") ON UPDATE RESTRICT ON DELETE CASCADE,
  CONSTRAINT "fk_cypher_starter" FOREIGN KEY ("started_by") REFERENCES "users" ("id") ON UPDATE RESTRICT ON DELETE RESTRICT
);
CREATE INDEX "idx_cypher_server" ON "cypher_sessions" ("server_id");
CREATE INDEX "idx_cypher_status" ON "cypher_sessions" ("status");
CREATE INDEX "fk_cypher_beat" ON "cypher_sessions" ("beat_id");
CREATE INDEX "fk_cypher_starter" ON "cypher_sessions" ("started_by");

-- Create "server_members" table
CREATE TABLE "server_members" (
  "id" BIGSERIAL NOT NULL,
  "server_id" BIGINT NOT NULL,
  "user_id" BIGINT NOT NULL,
  "joined_at" TIMESTAMP(6) NULL,
  "left_at" TIMESTAMP(6) NULL,
  "created_at" TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "uq_server_members" UNIQUE ("server_id", "user_id"),
  CONSTRAINT "fk_sm_server" FOREIGN KEY ("server_id") REFERENCES "servers" ("id") ON UPDATE RESTRICT ON DELETE CASCADE,
  CONSTRAINT "fk_sm_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE RESTRICT ON DELETE CASCADE
);
CREATE INDEX "idx_sm_server" ON "server_members" ("server_id");
CREATE INDEX "idx_sm_user" ON "server_members" ("user_id");

-- Create "tournaments" table
CREATE TABLE "tournaments" (
  "id" BIGSERIAL NOT NULL,
  "server_id" BIGINT NOT NULL,
  "thread_id" VARCHAR(32) NOT NULL,
  "name" VARCHAR(191) NOT NULL,
  "max_players" INTEGER NOT NULL,
  "status" tournament_status_type NOT NULL DEFAULT 'recruiting',
  "created_by" BIGINT NOT NULL,
  "started_at" TIMESTAMP(6) NULL,
  "finished_at" TIMESTAMP(6) NULL,
  "created_at" TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_tournaments_creator" FOREIGN KEY ("created_by") REFERENCES "users" ("id") ON UPDATE RESTRICT ON DELETE RESTRICT,
  CONSTRAINT "fk_tournaments_server" FOREIGN KEY ("server_id") REFERENCES "servers" ("id") ON UPDATE RESTRICT ON DELETE CASCADE
);
CREATE INDEX "idx_tournaments_server" ON "tournaments" ("server_id");
CREATE INDEX "idx_tournaments_thread" ON "tournaments" ("thread_id");
CREATE INDEX "fk_tournaments_creator" ON "tournaments" ("created_by");

-- Create "tournament_entries" table
CREATE TABLE "tournament_entries" (
  "id" BIGSERIAL NOT NULL,
  "tournament_id" BIGINT NOT NULL,
  "user_id" BIGINT NOT NULL,
  "is_active" BOOLEAN NOT NULL DEFAULT true,
  "joined_at" TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "left_at" TIMESTAMP(6) NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "uq_tournament_entry_active" UNIQUE ("tournament_id", "user_id", "is_active"),
  CONSTRAINT "fk_tournament_entries_tournament" FOREIGN KEY ("tournament_id") REFERENCES "tournaments" ("id") ON UPDATE RESTRICT ON DELETE CASCADE,
  CONSTRAINT "fk_tournament_entries_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE RESTRICT ON DELETE CASCADE
);
CREATE INDEX "idx_tournament_entries_tournament" ON "tournament_entries" ("tournament_id");
CREATE INDEX "idx_tournament_entries_user" ON "tournament_entries" ("user_id");

-- Create "tournament_matches" table
CREATE TABLE "tournament_matches" (
  "id" BIGSERIAL NOT NULL,
  "tournament_id" BIGINT NOT NULL,
  "round_no" INTEGER NOT NULL,
  "match_no" INTEGER NOT NULL,
  "player1_entry_id" BIGINT NOT NULL,
  "player2_entry_id" BIGINT NOT NULL,
  "beat_id" BIGINT NULL,
  "status" tournament_match_status_type NOT NULL DEFAULT 'pending',
  "winner_entry_id" BIGINT NULL,
  "started_at" TIMESTAMP(6) NULL,
  "voting_started_at" TIMESTAMP(6) NULL,
  "finished_at" TIMESTAMP(6) NULL,
  "created_at" TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "uq_match_in_tournament" UNIQUE ("tournament_id", "round_no", "match_no"),
  CONSTRAINT "fk_tm_beat" FOREIGN KEY ("beat_id") REFERENCES "beats" ("id") ON UPDATE RESTRICT ON DELETE SET NULL,
  CONSTRAINT "fk_tm_p1" FOREIGN KEY ("player1_entry_id") REFERENCES "tournament_entries" ("id") ON UPDATE RESTRICT ON DELETE RESTRICT,
  CONSTRAINT "fk_tm_p2" FOREIGN KEY ("player2_entry_id") REFERENCES "tournament_entries" ("id") ON UPDATE RESTRICT ON DELETE RESTRICT,
  CONSTRAINT "fk_tm_tournament" FOREIGN KEY ("tournament_id") REFERENCES "tournaments" ("id") ON UPDATE RESTRICT ON DELETE CASCADE,
  CONSTRAINT "fk_tm_winner" FOREIGN KEY ("winner_entry_id") REFERENCES "tournament_entries" ("id") ON UPDATE RESTRICT ON DELETE SET NULL,
  CONSTRAINT "chk_tm_distinct_players" CHECK ("player1_entry_id" <> "player2_entry_id")
);
CREATE INDEX "idx_matches_tournament" ON "tournament_matches" ("tournament_id");
CREATE INDEX "idx_matches_status" ON "tournament_matches" ("status");
CREATE INDEX "fk_tm_p1" ON "tournament_matches" ("player1_entry_id");
CREATE INDEX "fk_tm_p2" ON "tournament_matches" ("player2_entry_id");
CREATE INDEX "fk_tm_beat" ON "tournament_matches" ("beat_id");
CREATE INDEX "fk_tm_winner" ON "tournament_matches" ("winner_entry_id");

-- Create "tournament_votes" table
CREATE TABLE "tournament_votes" (
  "id" BIGSERIAL NOT NULL,
  "match_id" BIGINT NOT NULL,
  "voter_user_id" BIGINT NOT NULL,
  "voted_entry_id" BIGINT NOT NULL,
  "created_at" TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "uq_vote_unique" UNIQUE ("match_id", "voter_user_id"),
  CONSTRAINT "fk_votes_match" FOREIGN KEY ("match_id") REFERENCES "tournament_matches" ("id") ON UPDATE RESTRICT ON DELETE CASCADE,
  CONSTRAINT "fk_votes_voted_entry" FOREIGN KEY ("voted_entry_id") REFERENCES "tournament_entries" ("id") ON UPDATE RESTRICT ON DELETE RESTRICT,
  CONSTRAINT "fk_votes_voter" FOREIGN KEY ("voter_user_id") REFERENCES "users" ("id") ON UPDATE RESTRICT ON DELETE RESTRICT
);
CREATE INDEX "idx_votes_match" ON "tournament_votes" ("match_id");
CREATE INDEX "idx_votes_voted_entry" ON "tournament_votes" ("voted_entry_id");
CREATE INDEX "fk_votes_voter" ON "tournament_votes" ("voter_user_id");

-- Create "udon_clients" table
CREATE TABLE "udon_clients" (
  "id" BIGSERIAL NOT NULL,
  "server_id" BIGINT NOT NULL,
  "name" VARCHAR(191) NOT NULL,
  "token_hash" CHAR(64) NOT NULL,
  "is_active" BOOLEAN NOT NULL DEFAULT true,
  "last_access_at" TIMESTAMP(6) NULL,
  "last_access_ip" VARCHAR(45) NULL,
  "created_at" TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_udon_clients_server" FOREIGN KEY ("server_id") REFERENCES "servers" ("id") ON UPDATE RESTRICT ON DELETE CASCADE
);
CREATE INDEX "idx_udon_clients_server" ON "udon_clients" ("server_id");

-- Create "vrc_registrations" table
CREATE TABLE "vrc_registrations" (
  "id" BIGSERIAL NOT NULL,
  "server_id" BIGINT NOT NULL,
  "user_id" BIGINT NOT NULL,
  "vrc_name" VARCHAR(64) NOT NULL,
  "vrc_name_norm" VARCHAR(64) GENERATED ALWAYS AS (LOWER("vrc_name")) STORED,
  "status" vrc_registration_status_type NOT NULL DEFAULT 'pending',
  "reviewer_user_id" BIGINT NULL,
  "reviewed_at" TIMESTAMP(6) NULL,
  "note" VARCHAR(255) NULL,
  "created_at" TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "uq_vrc_unique_active" UNIQUE ("server_id", "vrc_name_norm", "status"),
  CONSTRAINT "fk_vrc_reviewer" FOREIGN KEY ("reviewer_user_id") REFERENCES "users" ("id") ON UPDATE RESTRICT ON DELETE SET NULL,
  CONSTRAINT "fk_vrc_server" FOREIGN KEY ("server_id") REFERENCES "servers" ("id") ON UPDATE RESTRICT ON DELETE CASCADE,
  CONSTRAINT "fk_vrc_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE RESTRICT ON DELETE CASCADE
);
CREATE INDEX "idx_vrc_by_user" ON "vrc_registrations" ("server_id", "user_id", "status");
CREATE INDEX "fk_vrc_user" ON "vrc_registrations" ("user_id");
CREATE INDEX "fk_vrc_reviewer" ON "vrc_registrations" ("reviewer_user_id");

-- Create "vrc_registration_events" table
CREATE TABLE "vrc_registration_events" (
  "id" BIGSERIAL NOT NULL,
  "registration_id" BIGINT NOT NULL,
  "event" vrc_event_type NOT NULL,
  "actor_user_id" BIGINT NULL,
  "detail" VARCHAR(255) NULL,
  "created_at" TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_vrc_ev_actor" FOREIGN KEY ("actor_user_id") REFERENCES "users" ("id") ON UPDATE RESTRICT ON DELETE SET NULL,
  CONSTRAINT "fk_vrc_ev_reg" FOREIGN KEY ("registration_id") REFERENCES "vrc_registrations" ("id") ON UPDATE RESTRICT ON DELETE CASCADE
);
CREATE INDEX "idx_vrc_ev_reg" ON "vrc_registration_events" ("registration_id");
CREATE INDEX "fk_vrc_ev_actor" ON "vrc_registration_events" ("actor_user_id");
