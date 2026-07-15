-- 流程实例编号迁移。
-- 前置要求：先清空流程运行数据。本迁移不兼容已有流程实例，不创建物理外键。
ALTER TABLE `wf_process_instance`
  ADD COLUMN `instance_no` VARCHAR(32) NOT NULL COMMENT '流程实例编号' AFTER `instance_id`,
  ADD UNIQUE KEY `uk_wf_process_instance_no` (`instance_no`);

CREATE TABLE `wf_process_instance_sequence` (
  `prefix` VARCHAR(8) NOT NULL COMMENT '流程分类前缀',
  `business_date` DATE NOT NULL COMMENT '业务日期',
  `current_value` INT UNSIGNED NOT NULL COMMENT '当前流水号',
  `create_date` DATETIME NOT NULL COMMENT '创建时间',
  `update_date` DATETIME NOT NULL COMMENT '更新时间',
  PRIMARY KEY (`prefix`, `business_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='流程实例编号每日流水';

-- 回滚方案（仅限已清空流程运行数据时执行）：
-- DROP TABLE `wf_process_instance_sequence`;
-- ALTER TABLE `wf_process_instance` DROP COLUMN `instance_no`;
