-- 流程表单结构迁移。
-- 兼容性：两个字段均允许 NULL，现有流程定义和实例不会被改写。
-- 说明：项目采用逻辑关联，不创建物理外键。

ALTER TABLE `wf_process_definition`
  ADD COLUMN `form_schema` LONGTEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL COMMENT '流程申请表单结构JSON' AFTER `flow_data`;
wo
ALTER TABLE `wf_process_instance`
  ADD COLUMN `form_snapshot` LONGTEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL COMMENT '发起时的流程表单结构快照JSON' AFTER `flow_snapshot`;

-- 回滚方案（执行前请先备份表单配置）：
-- ALTER TABLE `wf_process_instance` DROP COLUMN `form_snapshot`;
-- ALTER TABLE `wf_process_definition` DROP COLUMN `form_schema`;
