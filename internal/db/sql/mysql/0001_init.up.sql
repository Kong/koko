CREATE TABLE IF NOT EXISTS store (
  `key` VARCHAR(512) NOT NULL,
  value JSON,
  PRIMARY KEY (`key`),
  INDEX tags_idx ((CAST(value -> '$.object.tags' AS CHAR(128) ARRAY)))
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_520_ci;
