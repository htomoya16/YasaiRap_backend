CREATE VIEW v_discord_vrc_whitelist AS
SELECT
    w.id        AS whitelist_id,
    w.user_id   AS discord_id,
    wi.vrc_name AS vrc_name,
    wi.vrc_name_norm
FROM whitelists w
JOIN whitelist_items wi
  ON wi.whitelist_id = w.id
WHERE
    w.platform = 'discord';