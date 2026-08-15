-- 数据权限查询索引回滚。

ALTER TABLE `med_schedule`
  DROP INDEX `idx_med_schedule_owner_active`;

ALTER TABLE `med_registration`
  DROP INDEX `idx_med_registration_owner_date`;

ALTER TABLE `med_patient`
  DROP INDEX `idx_med_patient_owner_active`;

ALTER TABLE `erp_inventory_trace_code`
  DROP INDEX `idx_erp_inventory_trace_owner_date`;

ALTER TABLE `erp_inventory_movement`
  DROP INDEX `idx_erp_inventory_movement_owner_date`;

ALTER TABLE `erp_inventory_balance`
  DROP INDEX `idx_erp_inventory_balance_owner_date`;

ALTER TABLE `erp_other_outbound`
  DROP INDEX `idx_erp_other_outbound_owner_date`;

ALTER TABLE `erp_purchase_inbound`
  DROP INDEX `idx_erp_purchase_inbound_owner_date`;

ALTER TABLE `erp_purchase_order`
  DROP INDEX `idx_erp_purchase_order_owner_date`;

ALTER TABLE `dev_bug`
  DROP INDEX `idx_dev_bug_assignee_active`,
  DROP INDEX `idx_dev_bug_creator_active`;

ALTER TABLE `dev_task`
  DROP INDEX `idx_dev_task_assignee_active`,
  DROP INDEX `idx_dev_task_creator_active`;

ALTER TABLE `dev_story`
  DROP INDEX `idx_dev_story_owner_active`;

ALTER TABLE `dev_version`
  DROP INDEX `idx_dev_version_owner_active`;

ALTER TABLE `sys_login_log`
  DROP INDEX `idx_sys_login_log_owner_date`;

ALTER TABLE `sys_operation_log`
  DROP INDEX `idx_sys_operation_log_owner_date`;

ALTER TABLE `sys_file`
  DROP INDEX `idx_sys_file_owner_date`;

ALTER TABLE `sys_dept`
  DROP INDEX `idx_sys_dept_active_parent`;

ALTER TABLE `sys_user_dept`
  DROP INDEX `idx_sys_user_dept_dept_active`,
  DROP INDEX `idx_sys_user_dept_user_active`;

ALTER TABLE `sys_user_role`
  DROP INDEX `idx_sys_user_role_user_active`;
