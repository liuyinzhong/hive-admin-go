-- 数据权限查询索引。
-- 本脚本只增加二级索引，不改变业务数据；建议在低峰期执行并观察锁等待。

ALTER TABLE `sys_user_role`
  ADD INDEX `idx_sys_user_role_user_active` (`user_id`, `del_flag`, `role_id`);

ALTER TABLE `sys_user_dept`
  ADD INDEX `idx_sys_user_dept_user_active` (`user_id`, `del_flag`, `dept_id`),
  ADD INDEX `idx_sys_user_dept_dept_active` (`dept_id`, `del_flag`, `user_id`);

ALTER TABLE `sys_dept`
  ADD INDEX `idx_sys_dept_active_parent` (`del_flag`, `status`, `pid`, `dept_id`);

ALTER TABLE `sys_file`
  ADD INDEX `idx_sys_file_owner_date` (`creator_id`, `create_date`);

ALTER TABLE `sys_operation_log`
  ADD INDEX `idx_sys_operation_log_owner_date` (`user_id`, `create_date`);

ALTER TABLE `sys_login_log`
  ADD INDEX `idx_sys_login_log_owner_date` (`user_id`, `create_date`);

ALTER TABLE `dev_version`
  ADD INDEX `idx_dev_version_owner_active` (`creator_id`, `del_flag`);

ALTER TABLE `dev_story`
  ADD INDEX `idx_dev_story_owner_active` (`creator_id`, `del_flag`);

ALTER TABLE `dev_task`
  ADD INDEX `idx_dev_task_creator_active` (`creator_id`, `del_flag`),
  ADD INDEX `idx_dev_task_assignee_active` (`user_id`, `del_flag`);

ALTER TABLE `dev_bug`
  ADD INDEX `idx_dev_bug_creator_active` (`creator_id`, `del_flag`),
  ADD INDEX `idx_dev_bug_assignee_active` (`user_id`, `del_flag`);

ALTER TABLE `erp_purchase_order`
  ADD INDEX `idx_erp_purchase_order_owner_date` (`creator_id`, `create_date`);

ALTER TABLE `erp_purchase_inbound`
  ADD INDEX `idx_erp_purchase_inbound_owner_date` (`creator_id`, `create_date`);

ALTER TABLE `erp_other_outbound`
  ADD INDEX `idx_erp_other_outbound_owner_date` (`creator_id`, `create_date`);

ALTER TABLE `erp_inventory_balance`
  ADD INDEX `idx_erp_inventory_balance_owner_date` (`creator_id`, `update_date`);

ALTER TABLE `erp_inventory_movement`
  ADD INDEX `idx_erp_inventory_movement_owner_date` (`operator_id`, `create_date`);

ALTER TABLE `erp_inventory_trace_code`
  ADD INDEX `idx_erp_inventory_trace_owner_date` (`creator_id`, `update_date`);

ALTER TABLE `med_patient`
  ADD INDEX `idx_med_patient_owner_active` (`creator_id`, `del_flag`);

ALTER TABLE `med_registration`
  ADD INDEX `idx_med_registration_owner_date` (`creator_id`, `create_date`);

ALTER TABLE `med_schedule`
  ADD INDEX `idx_med_schedule_owner_active` (`creator_id`, `del_flag`);
