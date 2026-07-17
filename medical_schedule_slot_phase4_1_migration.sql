-- 医生出诊排班增量：半小时号源档位、发布时费用快照、自动任务记录和动态菜单。
-- 已执行的 medical_schedule_phase4_migration.sql 不回改，本脚本可重复执行。
-- 本脚本不包含 DELETE、DROP 或 TRUNCATE。

SET @sql = IF(
  (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'med_schedule_template' AND column_name = 'default_slot_quota') = 0,
  'ALTER TABLE `med_schedule_template` ADD COLUMN `default_slot_quota` INT NOT NULL DEFAULT 1 COMMENT ''每半小时默认号源容量'' AFTER `end_time`',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql = IF(
  (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'med_schedule_template' AND column_name = 'slot_quota_config') = 0,
  'ALTER TABLE `med_schedule_template` ADD COLUMN `slot_quota_config` LONGTEXT NULL COMMENT ''单档容量覆盖JSON'' AFTER `default_slot_quota`',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql = IF(
  (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'med_schedule' AND column_name = 'default_slot_quota') = 0,
  'ALTER TABLE `med_schedule` ADD COLUMN `default_slot_quota` INT NOT NULL DEFAULT 1 COMMENT ''草稿生成时每半小时默认容量'' AFTER `fee_amount`',
  'SELECT 1'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 草稿阶段允许费用为空，发布时由服务在事务中锁定有效规则并固化快照。
ALTER TABLE `med_schedule`
  MODIFY COLUMN `fee_rule_id` CHAR(36) NULL COMMENT '发布时固化的挂号费规则ID',
  MODIFY COLUMN `fee_rule_version` INT NULL COMMENT '发布时固化的规则版本',
  MODIFY COLUMN `fee_amount` DECIMAL(10,2) NULL COMMENT '发布时固化的挂号费金额（元）';

CREATE TABLE IF NOT EXISTS `med_schedule_slot` (
  `slot_id` CHAR(36) NOT NULL COMMENT '号源档位ID',
  `schedule_id` CHAR(36) NOT NULL COMMENT '出诊排班ID',
  `start_time` TIME NOT NULL COMMENT '档位开始时间',
  `end_time` TIME NOT NULL COMMENT '档位结束时间',
  `quota` INT NOT NULL COMMENT '档位可预约容量 0至99',
  `booked_quota` INT NOT NULL DEFAULT 0 COMMENT '已预约数量',
  `creator_id` CHAR(36) NULL COMMENT '创建人ID',
  `updater_id` CHAR(36) NULL COMMENT '更新人ID',
  `create_date` DATETIME NOT NULL COMMENT '创建时间',
  `update_date` DATETIME NOT NULL COMMENT '更新时间',
  `del_flag` TINYINT NOT NULL DEFAULT 0 COMMENT '删除标记 0正常 1删除',
  PRIMARY KEY (`slot_id`),
  UNIQUE KEY `uk_med_schedule_slot_time` (`schedule_id`, `start_time`, `del_flag`),
  KEY `idx_med_schedule_slot_available` (`schedule_id`, `start_time`, `quota`, `booked_quota`, `del_flag`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='出诊排班半小时预约号源档位';

CREATE TABLE IF NOT EXISTS `med_schedule_auto_task` (
  `task_id` CHAR(36) NOT NULL COMMENT '自动任务记录ID',
  `task_key` VARCHAR(96) NOT NULL COMMENT '任务幂等键：任务类型+目标周',
  `task_type` VARCHAR(16) NOT NULL COMMENT '任务类型 publish发布 generate生成',
  `target_week_start` DATE NOT NULL COMMENT '目标周开始日期（周一）',
  `target_week_end` DATE NOT NULL COMMENT '目标周结束日期（周日）',
  `status` TINYINT NOT NULL COMMENT '状态 0成功 1部分成功 2失败',
  `success_doctor_count` INT NOT NULL DEFAULT 0 COMMENT '成功医生数',
  `failure_doctor_count` INT NOT NULL DEFAULT 0 COMMENT '失败医生数',
  `details` LONGTEXT NULL COMMENT '失败医生及原因JSON',
  `executed_at` DATETIME NOT NULL COMMENT '执行时间',
  `create_date` DATETIME NOT NULL COMMENT '创建时间',
  `update_date` DATETIME NOT NULL COMMENT '更新时间',
  PRIMARY KEY (`task_id`),
  UNIQUE KEY `uk_med_schedule_auto_task_key` (`task_key`),
  KEY `idx_med_schedule_auto_task_week` (`target_week_start`, `task_type`, `status`),
  KEY `idx_med_schedule_auto_task_executed` (`executed_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='出诊排班每周自动发布和生成任务记录';

-- 动态菜单与操作权限。已有医疗管理权限的角色自动继承本次新增菜单和按钮。
SET @medical_catalog_id = (
  SELECT `id` FROM `sys_menu` WHERE `path` = '/medical' AND `del_flag` = 0 ORDER BY `create_date` LIMIT 1
);

INSERT INTO `sys_menu` (`id`, `pid`, `type`, `icon`, `title`, `name`, `path`, `auth_code`, `status`, `order`, `max_num_of_open_tab`, `create_date`, `update_date`, `del_flag`)
SELECT UUID(), @medical_catalog_id, 'catalog', 'lucide:calendar-clock', 'medical.schedule.title', 'MedicalSchedule', '/medical/schedule', 'medical:schedule:management', 1, 4, -1, NOW(), NOW(), 0
WHERE @medical_catalog_id IS NOT NULL AND NOT EXISTS (SELECT 1 FROM `sys_menu` WHERE `path` = '/medical/schedule' AND `del_flag` = 0);

SET @schedule_catalog_id = (SELECT `id` FROM `sys_menu` WHERE `path` = '/medical/schedule' AND `del_flag` = 0 ORDER BY `create_date` LIMIT 1);

INSERT INTO `sys_menu` (`id`, `pid`, `type`, `icon`, `component`, `title`, `name`, `path`, `auth_code`, `status`, `order`, `max_num_of_open_tab`, `create_date`, `update_date`, `del_flag`)
SELECT UUID(), @schedule_catalog_id, 'menu', 'lucide:calendar-days', '/medical/schedule/calendar/index', 'medical.schedule.calendarTitle', 'MedicalScheduleCalendar', '/medical/schedule/calendar', 'medical:schedule:list', 1, 1, -1, NOW(), NOW(), 0
WHERE @schedule_catalog_id IS NOT NULL AND NOT EXISTS (SELECT 1 FROM `sys_menu` WHERE `path` = '/medical/schedule/calendar' AND `del_flag` = 0);

INSERT INTO `sys_menu` (`id`, `pid`, `type`, `icon`, `component`, `title`, `name`, `path`, `auth_code`, `status`, `order`, `max_num_of_open_tab`, `create_date`, `update_date`, `del_flag`)
SELECT UUID(), @schedule_catalog_id, 'menu', 'lucide:calendar-sync', '/medical/schedule/template/list', 'medical.schedule.templateTitle', 'MedicalScheduleTemplate', '/medical/schedule/template', 'medical:schedule-template:list', 1, 2, -1, NOW(), NOW(), 0
WHERE @schedule_catalog_id IS NOT NULL AND NOT EXISTS (SELECT 1 FROM `sys_menu` WHERE `path` = '/medical/schedule/template' AND `del_flag` = 0);

INSERT INTO `sys_menu` (`id`, `pid`, `type`, `icon`, `component`, `title`, `name`, `path`, `auth_code`, `status`, `order`, `max_num_of_open_tab`, `create_date`, `update_date`, `del_flag`)
SELECT UUID(), @schedule_catalog_id, 'menu', 'lucide:history', '/medical/schedule/task/list', 'medical.schedule.taskTitle', 'MedicalScheduleTask', '/medical/schedule/task', 'medical:schedule-task:list', 1, 3, -1, NOW(), NOW(), 0
WHERE @schedule_catalog_id IS NOT NULL AND NOT EXISTS (SELECT 1 FROM `sys_menu` WHERE `path` = '/medical/schedule/task' AND `del_flag` = 0);

SET @schedule_calendar_menu_id = (SELECT `id` FROM `sys_menu` WHERE `path` = '/medical/schedule/calendar' AND `del_flag` = 0 ORDER BY `create_date` LIMIT 1);
SET @schedule_template_menu_id = (SELECT `id` FROM `sys_menu` WHERE `path` = '/medical/schedule/template' AND `del_flag` = 0 ORDER BY `create_date` LIMIT 1);

INSERT INTO `sys_menu` (`id`, `pid`, `type`, `title`, `name`, `auth_code`, `status`, `order`, `max_num_of_open_tab`, `create_date`, `update_date`, `del_flag`)
SELECT UUID(), @schedule_calendar_menu_id, 'button', 'medical.permission.scheduleCreate', 'MedicalScheduleCreate', 'medical:schedule:create', 1, 1, -1, NOW(), NOW(), 0
WHERE @schedule_calendar_menu_id IS NOT NULL AND NOT EXISTS (SELECT 1 FROM `sys_menu` WHERE `auth_code` = 'medical:schedule:create' AND `del_flag` = 0);
INSERT INTO `sys_menu` (`id`, `pid`, `type`, `title`, `name`, `auth_code`, `status`, `order`, `max_num_of_open_tab`, `create_date`, `update_date`, `del_flag`)
SELECT UUID(), @schedule_calendar_menu_id, 'button', 'medical.permission.scheduleUpdate', 'MedicalScheduleUpdate', 'medical:schedule:update', 1, 2, -1, NOW(), NOW(), 0
WHERE @schedule_calendar_menu_id IS NOT NULL AND NOT EXISTS (SELECT 1 FROM `sys_menu` WHERE `auth_code` = 'medical:schedule:update' AND `del_flag` = 0);
INSERT INTO `sys_menu` (`id`, `pid`, `type`, `title`, `name`, `auth_code`, `status`, `order`, `max_num_of_open_tab`, `create_date`, `update_date`, `del_flag`)
SELECT UUID(), @schedule_calendar_menu_id, 'button', 'medical.permission.scheduleDelete', 'MedicalScheduleDelete', 'medical:schedule:delete', 1, 3, -1, NOW(), NOW(), 0
WHERE @schedule_calendar_menu_id IS NOT NULL AND NOT EXISTS (SELECT 1 FROM `sys_menu` WHERE `auth_code` = 'medical:schedule:delete' AND `del_flag` = 0);
INSERT INTO `sys_menu` (`id`, `pid`, `type`, `title`, `name`, `auth_code`, `status`, `order`, `max_num_of_open_tab`, `create_date`, `update_date`, `del_flag`)
SELECT UUID(), @schedule_calendar_menu_id, 'button', 'medical.permission.scheduleGenerate', 'MedicalScheduleGenerate', 'medical:schedule:generate', 1, 4, -1, NOW(), NOW(), 0
WHERE @schedule_calendar_menu_id IS NOT NULL AND NOT EXISTS (SELECT 1 FROM `sys_menu` WHERE `auth_code` = 'medical:schedule:generate' AND `del_flag` = 0);
INSERT INTO `sys_menu` (`id`, `pid`, `type`, `title`, `name`, `auth_code`, `status`, `order`, `max_num_of_open_tab`, `create_date`, `update_date`, `del_flag`)
SELECT UUID(), @schedule_calendar_menu_id, 'button', 'medical.permission.schedulePublish', 'MedicalSchedulePublish', 'medical:schedule:publish', 1, 5, -1, NOW(), NOW(), 0
WHERE @schedule_calendar_menu_id IS NOT NULL AND NOT EXISTS (SELECT 1 FROM `sys_menu` WHERE `auth_code` = 'medical:schedule:publish' AND `del_flag` = 0);
INSERT INTO `sys_menu` (`id`, `pid`, `type`, `title`, `name`, `auth_code`, `status`, `order`, `max_num_of_open_tab`, `create_date`, `update_date`, `del_flag`)
SELECT UUID(), @schedule_calendar_menu_id, 'button', 'medical.permission.scheduleStop', 'MedicalScheduleStop', 'medical:schedule:stop', 1, 6, -1, NOW(), NOW(), 0
WHERE @schedule_calendar_menu_id IS NOT NULL AND NOT EXISTS (SELECT 1 FROM `sys_menu` WHERE `auth_code` = 'medical:schedule:stop' AND `del_flag` = 0);

INSERT INTO `sys_menu` (`id`, `pid`, `type`, `title`, `name`, `auth_code`, `status`, `order`, `max_num_of_open_tab`, `create_date`, `update_date`, `del_flag`)
SELECT UUID(), @schedule_template_menu_id, 'button', 'medical.permission.scheduleTemplateCreate', 'MedicalScheduleTemplateCreate', 'medical:schedule-template:create', 1, 1, -1, NOW(), NOW(), 0
WHERE @schedule_template_menu_id IS NOT NULL AND NOT EXISTS (SELECT 1 FROM `sys_menu` WHERE `auth_code` = 'medical:schedule-template:create' AND `del_flag` = 0);
INSERT INTO `sys_menu` (`id`, `pid`, `type`, `title`, `name`, `auth_code`, `status`, `order`, `max_num_of_open_tab`, `create_date`, `update_date`, `del_flag`)
SELECT UUID(), @schedule_template_menu_id, 'button', 'medical.permission.scheduleTemplateUpdate', 'MedicalScheduleTemplateUpdate', 'medical:schedule-template:update', 1, 2, -1, NOW(), NOW(), 0
WHERE @schedule_template_menu_id IS NOT NULL AND NOT EXISTS (SELECT 1 FROM `sys_menu` WHERE `auth_code` = 'medical:schedule-template:update' AND `del_flag` = 0);
INSERT INTO `sys_menu` (`id`, `pid`, `type`, `title`, `name`, `auth_code`, `status`, `order`, `max_num_of_open_tab`, `create_date`, `update_date`, `del_flag`)
SELECT UUID(), @schedule_template_menu_id, 'button', 'medical.permission.scheduleTemplateStatus', 'MedicalScheduleTemplateStatus', 'medical:schedule-template:status', 1, 3, -1, NOW(), NOW(), 0
WHERE @schedule_template_menu_id IS NOT NULL AND NOT EXISTS (SELECT 1 FROM `sys_menu` WHERE `auth_code` = 'medical:schedule-template:status' AND `del_flag` = 0);

-- 将本次新增菜单与按钮赋予已拥有“医疗管理”目录的角色，不扩大到其他角色。
INSERT INTO `sys_role_menu` (`id`, `role_id`, `menu_id`, `create_date`, `update_date`, `del_flag`)
SELECT UUID(), source_role.`role_id`, new_menu.`id`, NOW(), NOW(), 0
FROM (
  SELECT DISTINCT role_menu.`role_id`
  FROM `sys_role_menu` role_menu
  JOIN `sys_menu` medical_menu ON medical_menu.`id` = role_menu.`menu_id`
  WHERE medical_menu.`path` = '/medical' AND medical_menu.`del_flag` = 0 AND role_menu.`del_flag` = 0
) source_role
JOIN `sys_menu` new_menu ON (
  new_menu.`path` = '/medical/schedule'
  OR new_menu.`path` LIKE '/medical/schedule/%'
  OR new_menu.`auth_code` LIKE 'medical:schedule:%'
  OR new_menu.`auth_code` LIKE 'medical:schedule-template:%'
  OR new_menu.`auth_code` LIKE 'medical:schedule-task:%'
) AND new_menu.`del_flag` = 0
WHERE NOT EXISTS (
  SELECT 1 FROM `sys_role_menu` existing
  WHERE existing.`role_id` = source_role.`role_id` AND existing.`menu_id` = new_menu.`id` AND existing.`del_flag` = 0
);
