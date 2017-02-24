CREATE TABLE `chats` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT PRIMARY KEY,
  `chatid` bigint unsigned NOT NULL,
  `name` char(150) COLLATE 'utf8mb4_bin' NOT NULL,
  `messageTotal` int unsigned NOT NULL,
  `charTotal` bigint unsigned NOT NULL,
  `Type` char(30) COLLATE 'utf8mb4_bin' NOT NULL
) ENGINE='InnoDB' COLLATE 'utf8mb4_bin';

ALTER TABLE `chats`
ADD UNIQUE `chatid` (`chatid`);

CREATE TABLE `personstats` (
  `id` bigint NOT NULL AUTO_INCREMENT PRIMARY KEY,
  `chatfk` bigint(20) unsigned NOT NULL,
  `personid` bigint NOT NULL,
  `msgcount` bigint NOT NULL,
  `charcount` int NOT NULL,
  FOREIGN KEY (`chatfk`) REFERENCES `chats` (`chatid`) ON DELETE CASCADE
) ENGINE='InnoDB' COLLATE 'utf8mb4_bin';

ALTER TABLE `personstats`
ADD UNIQUE `personid` (`personid`);

ALTER TABLE `personstats`
ADD `name` varchar(64) COLLATE 'utf8mb4_bin' NOT NULL;