/*
Navicat MySQL Data Transfer

Source Server         : localhost
Source Server Version : 50161
Source Host           : localhost:3306
Source Database       : game_10

Target Server Type    : MYSQL
Target Server Version : 50161
File Encoding         : 65001

Date: 2018-08-15 16:08:21
*/

SET FOREIGN_KEY_CHECKS=0;
-- ----------------------------
-- Table structure for `account`
-- ----------------------------
DROP TABLE IF EXISTS `account`;
CREATE TABLE `account` (
  `accountid` int(11) NOT NULL,
  `accountname` char(255) NOT NULL,
  `password` char(255) NOT NULL,
  `createtime` datetime NOT NULL,
  `gmlevel` tinyint(4) NOT NULL,
  `data` text,
  PRIMARY KEY (`accountid`),
  KEY `index_account_name` (`accountname`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
-- Records of account
-- ----------------------------

-- ----------------------------
-- Table structure for `actor`
-- ----------------------------
DROP TABLE IF EXISTS `actor`;
CREATE TABLE `actor` (
  `actorid` bigint(20) NOT NULL,
  `actorname` char(30) NOT NULL,
  `accountid` int(11) NOT NULL,
  `serverid` int(11) NOT NULL,
  `camp` tinyint(4) NOT NULL,
  `sex` tinyint(4) NOT NULL,
  `level` int(11) NOT NULL,
  `power` int(11) NOT NULL,
  `createtime` datetime NOT NULL,
  `logintime` datetime NOT NULL,
  `logouttime` datetime NOT NULL,
  `basedata` text,
  `exdata` text,
  PRIMARY KEY (`actorid`),
  UNIQUE KEY `index_actor_name` (`actorname`),
  KEY `index_actor_accountid` (`accountid`),
  KEY `index_actor_serverid` (`serverid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
-- Records of actor
-- ----------------------------

-- ----------------------------
-- Table structure for `sysdata`
-- ----------------------------
DROP TABLE IF EXISTS `sysdata`;
CREATE TABLE `sysdata` (
  `id` tinyint(4) NOT NULL,
  `data` text,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
