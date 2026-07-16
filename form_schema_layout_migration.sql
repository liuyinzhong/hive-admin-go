-- 为独立表单增加响应式布局，并在流程实例中保存发起时的布局快照。
-- 前置要求：已执行 form_schema_refactor_migration.sql。

ALTER TABLE `sys_form_schema`
  ADD COLUMN `layout` VARCHAR(16) NOT NULL DEFAULT 'single' COMMENT '表单布局 single单列 double双列 triple三列' AFTER `category`;

ALTER TABLE `wf_process_instance`
  ADD COLUMN `form_layout` VARCHAR(16) NOT NULL DEFAULT 'single' COMMENT '发起时的表单布局快照' AFTER `form_snapshot`;

-- 回滚：
-- ALTER TABLE `wf_process_instance` DROP COLUMN `form_layout`;
-- ALTER TABLE `sys_form_schema` DROP COLUMN `layout`;
