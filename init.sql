CREATE TABLE `chats` (
  `chatid` bigint(20) NOT NULL,
  `name` char(150) COLLATE utf8mb4_bin NOT NULL,
  `messageTotal` int(10) unsigned NOT NULL,
  `charTotal` bigint(20) unsigned NOT NULL,
  `Type` char(30) COLLATE utf8mb4_bin NOT NULL,
  PRIMARY KEY (`chatid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

CREATE TABLE `personstats` (
  `chatfk` bigint(20) NOT NULL,
  `personid` bigint(20) NOT NULL,
  `msgcount` bigint(20) NOT NULL,
  `charcount` int(11) NOT NULL,
  `name` varchar(64) COLLATE utf8mb4_bin NOT NULL,
  PRIMARY KEY (`chatfk`,`personid`),
  CONSTRAINT `personstats_ibfk_1` FOREIGN KEY (`chatfk`) REFERENCES `chats` (`chatid`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;