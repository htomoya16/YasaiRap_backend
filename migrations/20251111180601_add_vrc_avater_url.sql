ALTER TABLE `whitelist_users`
  ADD COLUMN `vrc_avatar_url` varchar(512) NULL
    AFTER `vrc_display_name`;
ALTER TABLE `whitelist_users`
  CONVERT TO CHARACTER SET utf8mb4
  COLLATE utf8mb4_0900_ai_ci;