-- 医生管理第三阶段：挂号费规则、挂号类型字典、动态菜单和按钮权限。
-- 价格按“医生 + 出诊科室 + 挂号类型”独立版本化，历史金额不覆盖。
-- 项目使用逻辑关联，不创建物理外键。
-- 本脚本只包含创建、插入和必要的幂等更新；不包含 DELETE、DROP 或 TRUNCATE。

CREATE TABLE IF NOT EXISTS `med_registration_fee_rule` (
  `fee_rule_id` CHAR(36) NOT NULL COMMENT '挂号费规则ID',
  `doctor_id` CHAR(36) NOT NULL COMMENT '医生ID',
  `department_id` CHAR(36) NOT NULL COMMENT '出诊科室ID',
  `registration_type` VARCHAR(36) NOT NULL COMMENT '挂号类型，字典MED_REGISTRATION_TYPE',
  `fee_amount` DECIMAL(10,2) NOT NULL COMMENT '挂号费金额（元）',
  `effective_date` DATE NOT NULL COMMENT '生效日期（含）',
  `expiry_date` DATE NULL COMMENT '失效日期（含），NULL表示长期有效',
  `version` INT NOT NULL COMMENT '同一医生、科室和挂号类型下的版本号',
  `remark` VARCHAR(512) NULL COMMENT '调价说明或备注',
  `creator_id` CHAR(36) NULL COMMENT '创建人ID',
  `updater_id` CHAR(36) NULL COMMENT '更新人ID',
  `create_date` DATETIME NOT NULL COMMENT '创建时间',
  `update_date` DATETIME NOT NULL COMMENT '更新时间',
  `del_flag` TINYINT NOT NULL DEFAULT 0 COMMENT '删除标记 0正常 1删除',
  PRIMARY KEY (`fee_rule_id`),
  UNIQUE KEY `uk_med_registration_fee_version` (`doctor_id`, `department_id`, `registration_type`, `version`),
  KEY `idx_med_registration_fee_period` (`doctor_id`, `department_id`, `registration_type`, `effective_date`, `expiry_date`, `del_flag`),
  KEY `idx_med_registration_fee_department` (`department_id`, `effective_date`, `expiry_date`, `del_flag`),
  KEY `idx_med_registration_fee_effective` (`effective_date`, `expiry_date`, `del_flag`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='挂号费版本规则';

-- 挂号类型字典。
INSERT INTO `sys_dict` (`id`, `pid`, `type`, `label`, `value`, `remark`, `create_date`, `update_date`, `del_flag`, `status`)
SELECT UUID(), NULL, 'MED_REGISTRATION_TYPE', '挂号类型', NULL, '挂号费规则使用的挂号类型字典', NOW(), NOW(), 0, 1
WHERE NOT EXISTS (SELECT 1 FROM `sys_dict` WHERE `type` = 'MED_REGISTRATION_TYPE' AND `pid` IS NULL AND `del_flag` = 0);

SET @med_registration_type_dict_id = (
  SELECT `id` FROM `sys_dict`
  WHERE `type` = 'MED_REGISTRATION_TYPE' AND `pid` IS NULL AND `del_flag` = 0
  ORDER BY `create_date` LIMIT 1
);

INSERT INTO `sys_dict` (`id`, `pid`, `type`, `label`, `value`, `remark`, `create_date`, `update_date`, `del_flag`, `status`)
SELECT UUID(), @med_registration_type_dict_id, 'MED_REGISTRATION_TYPE', '普通', '1', NULL, NOW(), NOW(), 0, 1
WHERE NOT EXISTS (SELECT 1 FROM `sys_dict` WHERE `type` = 'MED_REGISTRATION_TYPE' AND `value` = '1' AND `del_flag` = 0);

INSERT INTO `sys_dict` (`id`, `pid`, `type`, `label`, `value`, `remark`, `create_date`, `update_date`, `del_flag`, `status`)
SELECT UUID(), @med_registration_type_dict_id, 'MED_REGISTRATION_TYPE', '专家', '2', NULL, NOW(), NOW(), 0, 1
WHERE NOT EXISTS (SELECT 1 FROM `sys_dict` WHERE `type` = 'MED_REGISTRATION_TYPE' AND `value` = '2' AND `del_flag` = 0);

INSERT INTO `sys_dict` (`id`, `pid`, `type`, `label`, `value`, `remark`, `create_date`, `update_date`, `del_flag`, `status`)
SELECT UUID(), @med_registration_type_dict_id, 'MED_REGISTRATION_TYPE', '特需', '3', NULL, NOW(), NOW(), 0, 1
WHERE NOT EXISTS (SELECT 1 FROM `sys_dict` WHERE `type` = 'MED_REGISTRATION_TYPE' AND `value` = '3' AND `del_flag` = 0);

INSERT INTO `sys_dict` (`id`, `pid`, `type`, `label`, `value`, `remark`, `create_date`, `update_date`, `del_flag`, `status`)
SELECT UUID(), @med_registration_type_dict_id, 'MED_REGISTRATION_TYPE', '复诊', '4', NULL, NOW(), NOW(), 0, 1
WHERE NOT EXISTS (SELECT 1 FROM `sys_dict` WHERE `type` = 'MED_REGISTRATION_TYPE' AND `value` = '4' AND `del_flag` = 0);

-- 动态菜单；不自动修改角色权限。
SET @medical_catalog_id = (
  SELECT `id` FROM `sys_menu`
  WHERE `path` = '/medical' AND `del_flag` = 0
  ORDER BY `create_date` LIMIT 1
);

INSERT INTO `sys_menu` (`id`, `pid`, `type`, `icon`, `component`, `title`, `name`, `path`, `auth_code`, `status`, `order`, `max_num_of_open_tab`, `create_date`, `update_date`, `del_flag`)
SELECT UUID(), @medical_catalog_id, 'menu', 'lucide:badge-dollar-sign', '/medical/registration-fee/list', 'medical.registrationFee.title', 'MedicalRegistrationFee', '/medical/registration-fee', 'medical:registration-fee:list', 1, 3, -1, NOW(), NOW(), 0
WHERE @medical_catalog_id IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM `sys_menu` WHERE `path` = '/medical/registration-fee' AND `del_flag` = 0);

SET @medical_registration_fee_menu_id = (
  SELECT `id` FROM `sys_menu`
  WHERE `path` = '/medical/registration-fee' AND `del_flag` = 0
  ORDER BY `create_date` LIMIT 1
);

INSERT INTO `sys_menu` (`id`, `pid`, `type`, `title`, `name`, `auth_code`, `status`, `order`, `max_num_of_open_tab`, `create_date`, `update_date`, `del_flag`)
SELECT UUID(), @medical_registration_fee_menu_id, 'button', 'medical.permission.registrationFeeCreate', 'MedicalRegistrationFeeCreate', 'medical:registration-fee:create', 1, 1, -1, NOW(), NOW(), 0
WHERE @medical_registration_fee_menu_id IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM `sys_menu` WHERE `auth_code` = 'medical:registration-fee:create' AND `del_flag` = 0);

INSERT INTO `sys_menu` (`id`, `pid`, `type`, `title`, `name`, `auth_code`, `status`, `order`, `max_num_of_open_tab`, `create_date`, `update_date`, `del_flag`)
SELECT UUID(), @medical_registration_fee_menu_id, 'button', 'medical.permission.registrationFeeAdjust', 'MedicalRegistrationFeeAdjust', 'medical:registration-fee:adjust', 1, 2, -1, NOW(), NOW(), 0
WHERE @medical_registration_fee_menu_id IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM `sys_menu` WHERE `auth_code` = 'medical:registration-fee:adjust' AND `del_flag` = 0);
