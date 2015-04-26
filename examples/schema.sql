-- Table for versioning table
CREATE TABLE version (
  id               BIGINT UNSIGNED	NOT NULL AUTO_INCREMENT,
  name        	   VARCHAR(255)   	NOT NULL UNIQUE,
  latest_version   BIGINT UNSIGNED 	NOT NULL DEFAULT 1,
  updated_at	   TIMESTAMP		NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY(id),
  KEY(name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- Table for register host table
CREATE TABLE host (
  id           BIGINT UNSIGNED	NOT NULL AUTO_INCREMENT,
  name	       VARCHAR(255)   	NOT NULL UNIQUE,
  target       VARCHAR(255)   	NOT NULL,
  key          VARCHAR(255)   	NOT NULL DEFAULT '',
  size         BIGINT UNSIGNED  NOT NULL DEFAULT 0,
  perent_id    BIGINT UNSIGNED 	NOT NULL,
  delegate     BOOLEAN          NOT NULL DEFAULT TRUE,
  created_at   TIMESTAMP	NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY(id),
  KEY(name),
  KEY(created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
