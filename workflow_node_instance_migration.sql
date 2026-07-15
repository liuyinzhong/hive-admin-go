-- 流程节点实例化迁移。
-- 前置要求：先清空 wf_process_instance、wf_process_task、wf_process_copy、wf_process_record 运行数据。
-- 本迁移不兼容旧运行数据，不创建物理外键。

CREATE TABLE `wf_process_node_instance` (
  `node_instance_id` CHAR(36) NOT NULL COMMENT '节点实例ID',
  `instance_id` CHAR(36) NOT NULL COMMENT '流程实例ID',
  `node_id` VARCHAR(128) NOT NULL COMMENT '流程节点ID',
  `node_name` VARCHAR(128) NOT NULL COMMENT '流程节点名称快照',
  `node_type` VARCHAR(32) NOT NULL COMMENT '节点类型',
  `sequence` INT NOT NULL COMMENT '实例内展示顺序',
  `route_version` INT NOT NULL COMMENT '路径版本',
  `status` TINYINT NOT NULL COMMENT '0预计 1处理中 2完成 3终止 4已替换',
  `approval_mode` VARCHAR(16) NULL COMMENT '审批方式 any/all',
  `branch_edge_id` VARCHAR(128) NULL COMMENT '条件节点命中连线ID',
  `actor_ids` LONGTEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '预计参与人ID数组JSON',
  `actor_names` LONGTEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '预计参与人名称数组JSON',
  `field_permissions` LONGTEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '节点字段权限快照JSON',
  `start_date` DATETIME NULL COMMENT '节点到达时间',
  `end_date` DATETIME NULL COMMENT '节点结束时间',
  `create_date` DATETIME NULL COMMENT '创建时间',
  `update_date` DATETIME NULL COMMENT '更新时间',
  PRIMARY KEY (`node_instance_id`),
  KEY `idx_wf_node_instance_instance_sequence` (`instance_id`, `sequence`),
  KEY `idx_wf_node_instance_instance_status` (`instance_id`, `status`),
  KEY `idx_wf_node_instance_route` (`instance_id`, `route_version`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='流程节点实例';

ALTER TABLE `wf_process_task`
  ADD COLUMN `node_instance_id` CHAR(36) NOT NULL COMMENT '节点实例ID' AFTER `task_group_id`,
  ADD KEY `idx_wf_task_node_instance` (`node_instance_id`);

ALTER TABLE `wf_process_copy`
  ADD COLUMN `node_instance_id` CHAR(36) NOT NULL COMMENT '节点实例ID' AFTER `copy_id`,
  ADD KEY `idx_wf_copy_node_instance` (`node_instance_id`);

ALTER TABLE `wf_process_record`
  ADD COLUMN `node_instance_id` CHAR(36) NOT NULL COMMENT '节点实例ID' AFTER `record_id`,
  ADD KEY `idx_wf_record_node_instance` (`node_instance_id`);

-- 回滚方案（仅限尚未产生新运行数据时执行）：
-- ALTER TABLE `wf_process_record` DROP COLUMN `node_instance_id`;
-- ALTER TABLE `wf_process_copy` DROP COLUMN `node_instance_id`;
-- ALTER TABLE `wf_process_task` DROP COLUMN `node_instance_id`;
-- DROP TABLE `wf_process_node_instance`;
