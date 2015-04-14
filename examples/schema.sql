-- Table for perent keys table
CREATE TABLE elton (
  id               BIGINT UNSIGNED	NOT NULL AUTO_INCREMENT,
  name        	   VARCHAR(255)   	NOT NULL,
  latest_version   BIGINT UNSIGNED 	NOT NULL DEFAULT 0,
  counter          BIGINT UNSIGNED 	NOT NULL DEFAULT 0,
  PRIMARY KEY(id),
  KEY(name)
) ENGINE=InnoDB CHARACTER SET 'utf8';

-- Table for elton table
CREATE TABLE elton_hosts (
  id         BIGINT UNSIGNED	NOT NULL AUTO_INCREMENT,
  name	     VARCHAR(255)   	NOT NULL,
  key        VARCHAR(255)   	NOT NULL,
  target     VARCHAR(255)   	NOT NULL,
  perent_id  BIGINT UNSIGNED 	NOT NULL,
  PRIMARY KEY(id),
  KEY(name),
  KEY(perent_id)
) ENGINE=InnoDB CHARACTER SET 'utf8';
