/*
 Navicat Premium Data Transfer

 Source Server         : 台式机
 Source Server Type    : MySQL
 Source Server Version : 50732
 Source Host           : 192.168.0.113:3306
 Source Schema         : bigbang

 Target Server Type    : MySQL
 Target Server Version : 50732
 File Encoding         : 65001

 Date: 05/01/2021 19:19:14
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for Vote
-- ----------------------------
DROP TABLE IF EXISTS `Vote`;
CREATE TABLE `Vote`  (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `addr` varchar(128) CHARACTER SET utf8 COLLATE utf8_bin NULL DEFAULT NULL,
  `a1` decimal(50, 20) NULL DEFAULT NULL,
  `a2` decimal(50, 20) NULL DEFAULT NULL,
  `v1` decimal(50, 20) NULL DEFAULT NULL,
  `v2` decimal(50, 20) NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8 COLLATE = utf8_bin ROW_FORMAT = Dynamic;

SET FOREIGN_KEY_CHECKS = 1;
