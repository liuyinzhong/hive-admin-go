CREATE TABLE IF NOT EXISTS `product_mp` (
  `mp_id` CHAR(36) NOT NULL COMMENT '厂家产品ID',
  `mp_code` VARCHAR(16) NOT NULL COMMENT 'MP编码',
  `rp_id` CHAR(36) NOT NULL COMMENT '所属规格产品ID',
  `enterprise_id` CHAR(36) NOT NULL COMMENT '生产企业ID',
  `approval_no` VARCHAR(128) NOT NULL COMMENT '批准文号/注册证号/备案号',
  `approval_no_normalized` VARCHAR(128) NOT NULL COMMENT '批准文号标准化值',
  `brand_name` VARCHAR(128) NULL COMMENT '品牌/商品名',
  `description` VARCHAR(2000) NULL COMMENT '描述',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态: 0停用 1启用',
  `row_version` INT NOT NULL DEFAULT 1 COMMENT '数据版本号',
  `creator_id` CHAR(36) NULL COMMENT '创建人ID',
  `updater_id` CHAR(36) NULL COMMENT '更新人ID',
  `create_date` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_date` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `del_flag` TINYINT NOT NULL DEFAULT 0 COMMENT '删除标记: 0正常 1删除',
  PRIMARY KEY (`mp_id`),
  UNIQUE KEY `uk_product_mp_code` (`mp_code`),
  UNIQUE KEY `uk_product_mp_rp_enterprise_approval` (`rp_id`, `enterprise_id`, `approval_no_normalized`, `del_flag`),
  KEY `idx_product_mp_rp_status` (`rp_id`, `status`),
  KEY `idx_product_mp_enterprise` (`enterprise_id`),
  KEY `idx_product_mp_update_date` (`update_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='厂家产品主数据表';

INSERT INTO `base_code_sequence` (`sequence_type`, `prefix`, `current_value`, `number_length`, `remark`, `update_date`)
SELECT 'PRODUCT_MP', 'MP', 0, 6, '厂家产品编码流水', NOW()
WHERE NOT EXISTS (
  SELECT 1 FROM `base_code_sequence` WHERE `sequence_type` = 'PRODUCT_MP'
);
