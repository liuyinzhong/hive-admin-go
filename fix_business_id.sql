-- 修改 dev_change_history 表的 business_id 字段长度
-- 将 char(12) 改为 char(36) 以支持 UUID 长度

ALTER TABLE `dev_change_history` 
MODIFY COLUMN `business_id` char(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '外键业务id,逻辑外键,根据关联类型决定是关联哪个表';
