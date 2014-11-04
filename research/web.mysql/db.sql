DROP DATABASE IF EXISTS srs_go;
CREATE DATABASE srs_go CHARACTER SET utf8 COLLATE utf8_general_ci;

USE srs_go;
SET NAMES 'utf8';

-- for srs servers.
DROP TABLE IF EXISTS `srs_server`;
CREATE TABLE `srs_server`(
    `server_id` int(32) NOT NULL AUTO_INCREMENT,
    `server_mac_addr` char(17) DEFAULT NULL COMMENT "the mac address of server",
    `server_ip_addr` varchar(39) DEFAULT NULL COMMENT "the ip v4/v6 address of server",
    `server_hostname` varchar(64) DEFAULT NULL COMMENT "the hostname of server",
    `server_create_time` datetime DEFAULT NULL COMMENT "the time server created in",
    `server_last_modify_time` datetime DEFAULT NULL COMMENT "the time server modify info",
    `server_last_heartbeat_time` datetime DEFAULT NULL COMMENT "the time server heartbeat",
    PRIMARY KEY (`server_id`)
) ENGINE=InnoDB AUTO_INCREMENT=100 DEFAULT CHARSET=utf8;

-- update history
-- 2014-11-14, create database.
