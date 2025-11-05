-- Create "atlas_schema_revisions" table
CREATE TABLE `atlas_schema_revisions` (
  `version` varchar(255) NOT NULL,
  `description` varchar(255) NOT NULL,
  `type` bigint unsigned NOT NULL DEFAULT 2,
  `applied` bigint NOT NULL DEFAULT 0,
  `total` bigint NOT NULL DEFAULT 0,
  `executed_at` timestamp NOT NULL,
  `execution_time` bigint NOT NULL,
  `error` longtext NULL,
  `error_stmt` longtext NULL,
  `hash` varchar(255) NOT NULL,
  `partial_hashes` json NULL,
  `operator_version` varchar(255) NOT NULL,
  PRIMARY KEY (`version`)
) CHARSET utf8mb4 COLLATE utf8mb4_bin;
-- Create "seasons" table
CREATE TABLE `seasons` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `slug` varchar(64) NOT NULL,
  `title` varchar(191) NOT NULL,
  `ruleset` enum('league','single_elim','double_elim') NOT NULL DEFAULT "league",
  `starts_on` date NULL,
  `ends_on` date NULL,
  `status` enum('draft','open','running','closed') NOT NULL DEFAULT "draft",
  `created_at` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  `updated_at` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (`id`),
  UNIQUE INDEX `uq_seasons_slug` (`slug`)
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- Create "users" table
CREATE TABLE `users` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `discord_id` varchar(32) NOT NULL,
  `name` varchar(191) NULL,
  `avatar_url` varchar(255) NULL,
  `created_at` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  `updated_at` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (`id`),
  UNIQUE INDEX `uq_users_discord` (`discord_id`)
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- Create "teams" table
CREATE TABLE `teams` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(191) NOT NULL,
  `owner_user_id` bigint unsigned NULL,
  `created_at` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  `updated_at` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (`id`),
  INDEX `fk_teams_owner` (`owner_user_id`),
  UNIQUE INDEX `uq_teams_name` (`name`),
  CONSTRAINT `fk_teams_owner` FOREIGN KEY (`owner_user_id`) REFERENCES `users` (`id`) ON UPDATE RESTRICT ON DELETE SET NULL
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- Create "entries" table
CREATE TABLE `entries` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `season_id` bigint unsigned NOT NULL,
  `kind` enum('solo','team') NOT NULL,
  `user_id` bigint unsigned NULL,
  `team_id` bigint unsigned NULL,
  `display_name` varchar(191) NULL,
  `created_at` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  `updated_at` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (`id`),
  INDEX `fk_entries_team` (`team_id`),
  INDEX `fk_entries_user` (`user_id`),
  INDEX `idx_entries_season` (`season_id`),
  UNIQUE INDEX `uq_entries_unique` (`season_id`, `kind`, `user_id`, `team_id`),
  CONSTRAINT `fk_entries_season` FOREIGN KEY (`season_id`) REFERENCES `seasons` (`id`) ON UPDATE RESTRICT ON DELETE CASCADE,
  CONSTRAINT `fk_entries_team` FOREIGN KEY (`team_id`) REFERENCES `teams` (`id`) ON UPDATE RESTRICT ON DELETE CASCADE,
  CONSTRAINT `fk_entries_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON UPDATE RESTRICT ON DELETE CASCADE,
  CONSTRAINT `chk_entries_exact_one` CHECK (((`kind` = _latin1'solo') and (`user_id` is not null) and (`team_id` is null)) or ((`kind` = _latin1'team') and (`team_id` is not null) and (`user_id` is null)))
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- Create "assets" table
CREATE TABLE `assets` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `kind` enum('standings','schedule','profile') NOT NULL,
  `season_id` bigint unsigned NULL,
  `entry_id` bigint unsigned NULL,
  `path` varchar(255) NOT NULL,
  `hash` char(40) NULL,
  `updated_at` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  `created_at` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  PRIMARY KEY (`id`),
  INDEX `idx_assets_entry` (`entry_id`, `kind`),
  INDEX `idx_assets_season` (`season_id`, `kind`),
  CONSTRAINT `fk_assets_entry` FOREIGN KEY (`entry_id`) REFERENCES `entries` (`id`) ON UPDATE RESTRICT ON DELETE CASCADE,
  CONSTRAINT `fk_assets_season` FOREIGN KEY (`season_id`) REFERENCES `seasons` (`id`) ON UPDATE RESTRICT ON DELETE CASCADE
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- Create "matches" table
CREATE TABLE `matches` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `season_id` bigint unsigned NOT NULL,
  `round_no` int unsigned NULL,
  `bracket` enum('main','losers') NULL,
  `home_entry_id` bigint unsigned NOT NULL,
  `away_entry_id` bigint unsigned NOT NULL,
  `scheduled_at` datetime(6) NULL,
  `status` enum('pending','scheduled','completed','canceled') NOT NULL DEFAULT "pending",
  `created_at` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  `updated_at` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (`id`),
  INDEX `fk_matches_away` (`away_entry_id`),
  INDEX `fk_matches_home` (`home_entry_id`),
  INDEX `idx_matches_season_round` (`season_id`, `round_no`),
  INDEX `idx_matches_time` (`season_id`, `scheduled_at`),
  CONSTRAINT `fk_matches_away` FOREIGN KEY (`away_entry_id`) REFERENCES `entries` (`id`) ON UPDATE RESTRICT ON DELETE CASCADE,
  CONSTRAINT `fk_matches_home` FOREIGN KEY (`home_entry_id`) REFERENCES `entries` (`id`) ON UPDATE RESTRICT ON DELETE CASCADE,
  CONSTRAINT `fk_matches_season` FOREIGN KEY (`season_id`) REFERENCES `seasons` (`id`) ON UPDATE RESTRICT ON DELETE CASCADE,
  CONSTRAINT `chk_matches_distinct` CHECK (`home_entry_id` <> `away_entry_id`)
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- Create "ratings" table
CREATE TABLE `ratings` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `season_id` bigint unsigned NOT NULL,
  `entry_id` bigint unsigned NOT NULL,
  `rating` double NOT NULL,
  `rd` double NULL,
  `computed_at` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  PRIMARY KEY (`id`),
  INDEX `fk_ratings_entry` (`entry_id`),
  INDEX `idx_ratings_latest` (`season_id`, `entry_id`, `computed_at`),
  UNIQUE INDEX `uq_ratings_point` (`season_id`, `entry_id`, `computed_at`),
  CONSTRAINT `fk_ratings_entry` FOREIGN KEY (`entry_id`) REFERENCES `entries` (`id`) ON UPDATE RESTRICT ON DELETE CASCADE,
  CONSTRAINT `fk_ratings_season` FOREIGN KEY (`season_id`) REFERENCES `seasons` (`id`) ON UPDATE RESTRICT ON DELETE CASCADE
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- Create "results" table
CREATE TABLE `results` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `match_id` bigint unsigned NOT NULL,
  `score_home` int NOT NULL,
  `score_away` int NOT NULL,
  `method` enum('normal','ff','dq') NOT NULL DEFAULT "normal",
  `reported_by` bigint unsigned NULL,
  `reported_at` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  `note` varchar(255) NULL,
  `created_at` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  `updated_at` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (`id`),
  INDEX `fk_results_reporter` (`reported_by`),
  UNIQUE INDEX `uq_results_match` (`match_id`),
  CONSTRAINT `fk_results_match` FOREIGN KEY (`match_id`) REFERENCES `matches` (`id`) ON UPDATE RESTRICT ON DELETE CASCADE,
  CONSTRAINT `fk_results_reporter` FOREIGN KEY (`reported_by`) REFERENCES `users` (`id`) ON UPDATE RESTRICT ON DELETE SET NULL
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- Create "servers" table
CREATE TABLE `servers` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `discord_guild_id` varchar(32) NOT NULL,
  `name` varchar(191) NULL,
  `icon_url` varchar(255) NULL,
  `owner_discord_id` varchar(32) NULL,
  `created_at` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  `updated_at` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (`id`),
  UNIQUE INDEX `uq_servers_discord` (`discord_guild_id`)
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- Create "server_members" table
CREATE TABLE `server_members` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `server_id` bigint unsigned NOT NULL,
  `user_id` bigint unsigned NOT NULL,
  `joined_at` timestamp(6) NULL,
  `left_at` timestamp(6) NULL,
  `created_at` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  `updated_at` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (`id`),
  INDEX `idx_sm_server` (`server_id`),
  INDEX `idx_sm_user` (`user_id`),
  UNIQUE INDEX `uq_server_members` (`server_id`, `user_id`),
  CONSTRAINT `fk_sm_server` FOREIGN KEY (`server_id`) REFERENCES `servers` (`id`) ON UPDATE RESTRICT ON DELETE CASCADE,
  CONSTRAINT `fk_sm_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON UPDATE RESTRICT ON DELETE CASCADE
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- Create "udon_clients" table
CREATE TABLE `udon_clients` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `server_id` bigint unsigned NOT NULL,
  `name` varchar(191) NOT NULL,
  `token_hash` char(64) NOT NULL,
  `is_active` bool NOT NULL DEFAULT 1,
  `last_access_at` timestamp(6) NULL,
  `last_access_ip` varchar(45) NULL,
  `created_at` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  `updated_at` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (`id`),
  INDEX `idx_udon_clients_server` (`server_id`),
  CONSTRAINT `fk_udon_clients_server` FOREIGN KEY (`server_id`) REFERENCES `servers` (`id`) ON UPDATE RESTRICT ON DELETE CASCADE
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- Create "vrc_registrations" table
CREATE TABLE `vrc_registrations` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `server_id` bigint unsigned NOT NULL,
  `user_id` bigint unsigned NOT NULL,
  `vrc_name` varchar(64) NOT NULL,
  `vrc_name_norm` varchar(64) AS (lower(`vrc_name`)) STORED NULL,
  `status` enum('pending','approved','rejected') NOT NULL DEFAULT "pending",
  `reviewer_user_id` bigint unsigned NULL,
  `reviewed_at` timestamp(6) NULL,
  `note` varchar(255) NULL,
  `created_at` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  `updated_at` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (`id`),
  INDEX `fk_vrc_reviewer` (`reviewer_user_id`),
  INDEX `fk_vrc_user` (`user_id`),
  INDEX `idx_vrc_by_user` (`server_id`, `user_id`, `status`),
  UNIQUE INDEX `uq_vrc_unique_active` (`server_id`, `vrc_name_norm`, `status`),
  CONSTRAINT `fk_vrc_reviewer` FOREIGN KEY (`reviewer_user_id`) REFERENCES `users` (`id`) ON UPDATE RESTRICT ON DELETE SET NULL,
  CONSTRAINT `fk_vrc_server` FOREIGN KEY (`server_id`) REFERENCES `servers` (`id`) ON UPDATE RESTRICT ON DELETE CASCADE,
  CONSTRAINT `fk_vrc_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON UPDATE RESTRICT ON DELETE CASCADE,
  CONSTRAINT `chk_vrc_status` CHECK (`status` in (_utf8mb4'pending',_utf8mb4'approved',_utf8mb4'rejected'))
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- Create "vrc_registration_events" table
CREATE TABLE `vrc_registration_events` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `registration_id` bigint unsigned NOT NULL,
  `event` enum('submitted','approved','rejected','revoked') NOT NULL,
  `actor_user_id` bigint unsigned NULL,
  `detail` varchar(255) NULL,
  `created_at` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  PRIMARY KEY (`id`),
  INDEX `fk_vrc_ev_actor` (`actor_user_id`),
  INDEX `idx_vrc_ev_reg` (`registration_id`),
  CONSTRAINT `fk_vrc_ev_actor` FOREIGN KEY (`actor_user_id`) REFERENCES `users` (`id`) ON UPDATE RESTRICT ON DELETE SET NULL,
  CONSTRAINT `fk_vrc_ev_reg` FOREIGN KEY (`registration_id`) REFERENCES `vrc_registrations` (`id`) ON UPDATE RESTRICT ON DELETE CASCADE
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- Create "whitelists" table
CREATE TABLE `whitelists` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `server_id` bigint unsigned NOT NULL,
  `version` bigint unsigned NOT NULL,
  `item_count` int unsigned NOT NULL DEFAULT 0,
  `path` varchar(255) NOT NULL,
  `content_hash` char(64) NOT NULL,
  `generated_by` bigint unsigned NULL,
  `generated_at` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  `expires_at` timestamp(6) NULL,
  `is_active` bool NOT NULL DEFAULT 1,
  PRIMARY KEY (`id`),
  INDEX `fk_whitelists_generator` (`generated_by`),
  INDEX `idx_whitelists_active` (`server_id`, `is_active`),
  UNIQUE INDEX `uq_whitelists_ver` (`server_id`, `version`),
  CONSTRAINT `fk_whitelists_generator` FOREIGN KEY (`generated_by`) REFERENCES `users` (`id`) ON UPDATE RESTRICT ON DELETE SET NULL,
  CONSTRAINT `fk_whitelists_server` FOREIGN KEY (`server_id`) REFERENCES `servers` (`id`) ON UPDATE RESTRICT ON DELETE CASCADE
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- Create "whitelist_items" table
CREATE TABLE `whitelist_items` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `whitelist_id` bigint unsigned NOT NULL,
  `vrc_name` varchar(64) NOT NULL,
  `vrc_name_norm` varchar(64) AS (lower(`vrc_name`)) STORED NULL,
  PRIMARY KEY (`id`),
  INDEX `idx_wli_norm` (`vrc_name_norm`),
  UNIQUE INDEX `uq_whitelist_item` (`whitelist_id`, `vrc_name_norm`),
  CONSTRAINT `fk_wli_whitelist` FOREIGN KEY (`whitelist_id`) REFERENCES `whitelists` (`id`) ON UPDATE RESTRICT ON DELETE CASCADE
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
