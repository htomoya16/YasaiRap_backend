-- Create "whitelist_users" table
CREATE TABLE `whitelist_users` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `discord_user_id` varchar(64) NOT NULL,
  `vrc_user_id` varchar(64) NOT NULL,
  `vrc_display_name` varchar(64) NOT NULL,
  `note` varchar(255) NOT NULL DEFAULT "",
  `created_at` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  `updated_at` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (`id`),
  UNIQUE INDEX `uq_discord_user` (`discord_user_id`),
  UNIQUE INDEX `uq_vrc_user` (`vrc_user_id`)
) CHARSET utf8mb4 COLLATE utf8mb4_unicode_ci;
-- Drop "whitelist_items" table
DROP TABLE `whitelist_items`;
-- Drop "whitelists" table
DROP TABLE `whitelists`;
DROP VIEW `v_discord_vrc_whitelist`;