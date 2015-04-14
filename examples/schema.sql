-- Table for perent keys table
CREATE TABLE version (
  id               BIGINT UNSIGNED	NOT NULL AUTO_INCREMENT,
  name        	   VARCHAR(255)   	NOT NULL UNIQUE,
  latest_version   BIGINT UNSIGNED 	NOT NULL DEFAULT 0,
  counter          BIGINT UNSIGNED 	NOT NULL DEFAULT 1,
  PRIMARY KEY(id),
  KEY(name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- Table for elton table
CREATE TABLE host (
  id         BIGINT UNSIGNED	NOT NULL AUTO_INCREMENT,
  name	     VARCHAR(255)   	NOT NULL UNIQUE,
  key        VARCHAR(255)   	NOT NULL,
  target     VARCHAR(255)   	NOT NULL,
  perent_id  BIGINT UNSIGNED 	NOT NULL,
  PRIMARY KEY(id),
  KEY(name),
  KEY(perent_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
