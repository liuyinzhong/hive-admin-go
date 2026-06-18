-- 用户表新增手机号字段
ALTER TABLE `sys_user`
ADD COLUMN `phone` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT '手机号' AFTER `email`;
