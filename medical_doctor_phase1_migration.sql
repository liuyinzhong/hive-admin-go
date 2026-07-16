-- 医生管理第一阶段：临床科室、医生基础档案、医生出诊科室、医疗字典与动态菜单。
-- 项目沿用逻辑关联，不创建物理外键。
-- 本脚本只包含创建和插入操作，可重复执行；不包含 DELETE、DROP 或 TRUNCATE。
-- Windows 环境请使用 MySQL SOURCE 命令直接读取本 UTF-8 文件，不要通过 PowerShell 文本管道传输。

CREATE TABLE IF NOT EXISTS `med_department` (
  `department_id` CHAR(36) NOT NULL COMMENT '临床科室ID',
  `department_code` VARCHAR(32) NOT NULL COMMENT '临床科室编码',
  `department_name` VARCHAR(64) NOT NULL COMMENT '临床科室名称',
  `pid` CHAR(36) NULL COMMENT '上级临床科室ID',
  `sort` INT NOT NULL DEFAULT 0 COMMENT '排序，升序',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 0停用 1启用',
  `remark` VARCHAR(512) NULL COMMENT '备注',
  `creator_id` CHAR(36) NULL COMMENT '创建人ID',
  `updater_id` CHAR(36) NULL COMMENT '更新人ID',
  `create_date` DATETIME NOT NULL COMMENT '创建时间',
  `update_date` DATETIME NOT NULL COMMENT '更新时间',
  `del_flag` TINYINT NOT NULL DEFAULT 0 COMMENT '删除标记 0正常 1删除',
  PRIMARY KEY (`department_id`),
  UNIQUE KEY `uk_med_department_code` (`department_code`),
  KEY `idx_med_department_pid` (`pid`, `del_flag`),
  KEY `idx_med_department_name` (`department_name`),
  KEY `idx_med_department_status` (`status`, `del_flag`, `sort`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='临床科室';

CREATE TABLE IF NOT EXISTS `med_doctor` (
  `doctor_id` CHAR(36) NOT NULL COMMENT '医生ID',
  `doctor_no` VARCHAR(32) NOT NULL COMMENT '医生编号',
  `user_id` CHAR(36) NULL COMMENT '绑定系统用户ID',
  `name` VARCHAR(64) NOT NULL COMMENT '医生姓名',
  `name_pinyin` VARCHAR(128) NULL COMMENT '姓名拼音',
  `gender` VARCHAR(36) NULL COMMENT '性别，字典MED_DOCTOR_GENDER',
  `birth_date` DATE NULL COMMENT '出生日期',
  `phone` VARCHAR(20) NULL COMMENT '工作联系电话',
  `email` VARCHAR(128) NULL COMMENT '工作邮箱',
  `avatar` VARCHAR(512) NULL COMMENT '头像URL',
  `professional_title` VARCHAR(36) NOT NULL COMMENT '职称，字典MED_DOCTOR_TITLE',
  `administrative_position` VARCHAR(64) NULL COMMENT '行政职务',
  `employment_type` VARCHAR(36) NOT NULL COMMENT '用工类型，字典MED_EMPLOYMENT_TYPE',
  `practice_start_date` DATE NULL COMMENT '开始从业日期',
  `employment_date` DATE NULL COMMENT '入职日期',
  `departure_date` DATE NULL COMMENT '离职日期',
  `expertise` TEXT NULL COMMENT '擅长领域',
  `introduction` TEXT NULL COMMENT '医生简介',
  `default_visit_minutes` SMALLINT NOT NULL DEFAULT 15 COMMENT '默认接诊分钟数',
  `online_consultation` TINYINT NOT NULL DEFAULT 0 COMMENT '是否支持线上问诊 0否 1是',
  `appointment_enabled` TINYINT NOT NULL DEFAULT 1 COMMENT '是否允许预约 0否 1是',
  `profile_visible` TINYINT NOT NULL DEFAULT 1 COMMENT '是否公开展示 0否 1是',
  `sort` INT NOT NULL DEFAULT 0 COMMENT '展示排序，升序',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 0停用 1启用',
  `remark` VARCHAR(512) NULL COMMENT '内部备注',
  `creator_id` CHAR(36) NULL COMMENT '创建人ID',
  `updater_id` CHAR(36) NULL COMMENT '更新人ID',
  `create_date` DATETIME NOT NULL COMMENT '创建时间',
  `update_date` DATETIME NOT NULL COMMENT '更新时间',
  `del_flag` TINYINT NOT NULL DEFAULT 0 COMMENT '删除标记 0正常 1删除',
  PRIMARY KEY (`doctor_id`),
  UNIQUE KEY `uk_med_doctor_no` (`doctor_no`),
  UNIQUE KEY `uk_med_doctor_user` (`user_id`),
  KEY `idx_med_doctor_name` (`name`),
  KEY `idx_med_doctor_title` (`professional_title`),
  KEY `idx_med_doctor_status` (`status`, `del_flag`, `sort`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='医生基础档案';

CREATE TABLE IF NOT EXISTS `med_doctor_department` (
  `doctor_department_id` CHAR(36) NOT NULL COMMENT '医生出诊科室关系ID',
  `doctor_id` CHAR(36) NOT NULL COMMENT '医生ID',
  `department_id` CHAR(36) NOT NULL COMMENT '临床科室ID',
  `is_primary` TINYINT NOT NULL DEFAULT 0 COMMENT '是否主科室 0否 1是',
  `department_position` VARCHAR(64) NULL COMMENT '医生在该科室的职务',
  `appointment_enabled` TINYINT NOT NULL DEFAULT 1 COMMENT '该科室是否允许预约 0否 1是',
  `valid_from` DATE NULL COMMENT '关系生效日期',
  `valid_to` DATE NULL COMMENT '关系失效日期',
  `sort` INT NOT NULL DEFAULT 0 COMMENT '排序，升序',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 0停用 1启用',
  `creator_id` CHAR(36) NULL COMMENT '创建人ID',
  `updater_id` CHAR(36) NULL COMMENT '更新人ID',
  `create_date` DATETIME NOT NULL COMMENT '创建时间',
  `update_date` DATETIME NOT NULL COMMENT '更新时间',
  `del_flag` TINYINT NOT NULL DEFAULT 0 COMMENT '删除标记 0正常 1删除',
  PRIMARY KEY (`doctor_department_id`),
  UNIQUE KEY `uk_med_doctor_department` (`doctor_id`, `department_id`),
  KEY `idx_med_doctor_department_dept` (`department_id`, `status`, `del_flag`),
  KEY `idx_med_doctor_department_primary` (`doctor_id`, `is_primary`, `status`, `del_flag`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='医生出诊科室';

-- 医疗字典：使用现有 sys_dict 的“根节点 + 子项”结构。
INSERT INTO `sys_dict` (`id`, `pid`, `type`, `label`, `value`, `remark`, `create_date`, `update_date`, `del_flag`, `status`)
SELECT UUID(), NULL, 'MED_DOCTOR_GENDER', '医生性别', NULL, '医生档案性别字典', NOW(), NOW(), 0, 1
WHERE NOT EXISTS (SELECT 1 FROM `sys_dict` WHERE `type` = 'MED_DOCTOR_GENDER' AND `pid` IS NULL AND `del_flag` = 0);
SET @med_gender_dict_id = (SELECT `id` FROM `sys_dict` WHERE `type` = 'MED_DOCTOR_GENDER' AND `pid` IS NULL AND `del_flag` = 0 ORDER BY `create_date` LIMIT 1);
INSERT INTO `sys_dict` (`id`, `pid`, `type`, `label`, `value`, `remark`, `create_date`, `update_date`, `del_flag`, `status`)
SELECT UUID(), @med_gender_dict_id, 'MED_DOCTOR_GENDER', '未知', '0', NULL, NOW(), NOW(), 0, 1
WHERE NOT EXISTS (SELECT 1 FROM `sys_dict` WHERE `type` = 'MED_DOCTOR_GENDER' AND `value` = '0' AND `del_flag` = 0);
INSERT INTO `sys_dict` (`id`, `pid`, `type`, `label`, `value`, `remark`, `create_date`, `update_date`, `del_flag`, `status`)
SELECT UUID(), @med_gender_dict_id, 'MED_DOCTOR_GENDER', '男', '1', NULL, NOW(), NOW(), 0, 1
WHERE NOT EXISTS (SELECT 1 FROM `sys_dict` WHERE `type` = 'MED_DOCTOR_GENDER' AND `value` = '1' AND `del_flag` = 0);
INSERT INTO `sys_dict` (`id`, `pid`, `type`, `label`, `value`, `remark`, `create_date`, `update_date`, `del_flag`, `status`)
SELECT UUID(), @med_gender_dict_id, 'MED_DOCTOR_GENDER', '女', '2', NULL, NOW(), NOW(), 0, 1
WHERE NOT EXISTS (SELECT 1 FROM `sys_dict` WHERE `type` = 'MED_DOCTOR_GENDER' AND `value` = '2' AND `del_flag` = 0);

INSERT INTO `sys_dict` (`id`, `pid`, `type`, `label`, `value`, `remark`, `create_date`, `update_date`, `del_flag`, `status`)
SELECT UUID(), NULL, 'MED_DOCTOR_TITLE', '医生职称', NULL, '医生专业技术职称字典', NOW(), NOW(), 0, 1
WHERE NOT EXISTS (SELECT 1 FROM `sys_dict` WHERE `type` = 'MED_DOCTOR_TITLE' AND `pid` IS NULL AND `del_flag` = 0);
SET @med_title_dict_id = (SELECT `id` FROM `sys_dict` WHERE `type` = 'MED_DOCTOR_TITLE' AND `pid` IS NULL AND `del_flag` = 0 ORDER BY `create_date` LIMIT 1);
INSERT INTO `sys_dict` (`id`, `pid`, `type`, `label`, `value`, `remark`, `create_date`, `update_date`, `del_flag`, `status`)
SELECT UUID(), @med_title_dict_id, 'MED_DOCTOR_TITLE', '住院医师', '1', NULL, NOW(), NOW(), 0, 1 WHERE NOT EXISTS (SELECT 1 FROM `sys_dict` WHERE `type` = 'MED_DOCTOR_TITLE' AND `value` = '1' AND `del_flag` = 0);
INSERT INTO `sys_dict` (`id`, `pid`, `type`, `label`, `value`, `remark`, `create_date`, `update_date`, `del_flag`, `status`)
SELECT UUID(), @med_title_dict_id, 'MED_DOCTOR_TITLE', '主治医师', '2', NULL, NOW(), NOW(), 0, 1 WHERE NOT EXISTS (SELECT 1 FROM `sys_dict` WHERE `type` = 'MED_DOCTOR_TITLE' AND `value` = '2' AND `del_flag` = 0);
INSERT INTO `sys_dict` (`id`, `pid`, `type`, `label`, `value`, `remark`, `create_date`, `update_date`, `del_flag`, `status`)
SELECT UUID(), @med_title_dict_id, 'MED_DOCTOR_TITLE', '副主任医师', '3', NULL, NOW(), NOW(), 0, 1 WHERE NOT EXISTS (SELECT 1 FROM `sys_dict` WHERE `type` = 'MED_DOCTOR_TITLE' AND `value` = '3' AND `del_flag` = 0);
INSERT INTO `sys_dict` (`id`, `pid`, `type`, `label`, `value`, `remark`, `create_date`, `update_date`, `del_flag`, `status`)
SELECT UUID(), @med_title_dict_id, 'MED_DOCTOR_TITLE', '主任医师', '4', NULL, NOW(), NOW(), 0, 1 WHERE NOT EXISTS (SELECT 1 FROM `sys_dict` WHERE `type` = 'MED_DOCTOR_TITLE' AND `value` = '4' AND `del_flag` = 0);
INSERT INTO `sys_dict` (`id`, `pid`, `type`, `label`, `value`, `remark`, `create_date`, `update_date`, `del_flag`, `status`)
SELECT UUID(), @med_title_dict_id, 'MED_DOCTOR_TITLE', '其他', '5', NULL, NOW(), NOW(), 0, 1 WHERE NOT EXISTS (SELECT 1 FROM `sys_dict` WHERE `type` = 'MED_DOCTOR_TITLE' AND `value` = '5' AND `del_flag` = 0);

INSERT INTO `sys_dict` (`id`, `pid`, `type`, `label`, `value`, `remark`, `create_date`, `update_date`, `del_flag`, `status`)
SELECT UUID(), NULL, 'MED_EMPLOYMENT_TYPE', '医生用工类型', NULL, '医生档案用工类型字典', NOW(), NOW(), 0, 1
WHERE NOT EXISTS (SELECT 1 FROM `sys_dict` WHERE `type` = 'MED_EMPLOYMENT_TYPE' AND `pid` IS NULL AND `del_flag` = 0);
SET @med_employment_dict_id = (SELECT `id` FROM `sys_dict` WHERE `type` = 'MED_EMPLOYMENT_TYPE' AND `pid` IS NULL AND `del_flag` = 0 ORDER BY `create_date` LIMIT 1);
INSERT INTO `sys_dict` (`id`, `pid`, `type`, `label`, `value`, `remark`, `create_date`, `update_date`, `del_flag`, `status`)
SELECT UUID(), @med_employment_dict_id, 'MED_EMPLOYMENT_TYPE', '全职', '1', NULL, NOW(), NOW(), 0, 1 WHERE NOT EXISTS (SELECT 1 FROM `sys_dict` WHERE `type` = 'MED_EMPLOYMENT_TYPE' AND `value` = '1' AND `del_flag` = 0);
INSERT INTO `sys_dict` (`id`, `pid`, `type`, `label`, `value`, `remark`, `create_date`, `update_date`, `del_flag`, `status`)
SELECT UUID(), @med_employment_dict_id, 'MED_EMPLOYMENT_TYPE', '兼职', '2', NULL, NOW(), NOW(), 0, 1 WHERE NOT EXISTS (SELECT 1 FROM `sys_dict` WHERE `type` = 'MED_EMPLOYMENT_TYPE' AND `value` = '2' AND `del_flag` = 0);
INSERT INTO `sys_dict` (`id`, `pid`, `type`, `label`, `value`, `remark`, `create_date`, `update_date`, `del_flag`, `status`)
SELECT UUID(), @med_employment_dict_id, 'MED_EMPLOYMENT_TYPE', '外聘', '3', NULL, NOW(), NOW(), 0, 1 WHERE NOT EXISTS (SELECT 1 FROM `sys_dict` WHERE `type` = 'MED_EMPLOYMENT_TYPE' AND `value` = '3' AND `del_flag` = 0);
INSERT INTO `sys_dict` (`id`, `pid`, `type`, `label`, `value`, `remark`, `create_date`, `update_date`, `del_flag`, `status`)
SELECT UUID(), @med_employment_dict_id, 'MED_EMPLOYMENT_TYPE', '多点执业', '4', NULL, NOW(), NOW(), 0, 1 WHERE NOT EXISTS (SELECT 1 FROM `sys_dict` WHERE `type` = 'MED_EMPLOYMENT_TYPE' AND `value` = '4' AND `del_flag` = 0);

-- 动态菜单；不自动修改角色权限。
INSERT INTO `sys_menu` (`id`, `pid`, `type`, `icon`, `title`, `name`, `path`, `auth_code`, `status`, `order`, `max_num_of_open_tab`, `create_date`, `update_date`, `del_flag`)
SELECT UUID(), NULL, 'catalog', 'lucide:stethoscope', 'medical.title', 'MedicalManagement', '/medical', 'medical:management', 1, 30, -1, NOW(), NOW(), 0
WHERE NOT EXISTS (SELECT 1 FROM `sys_menu` WHERE `path` = '/medical' AND `del_flag` = 0);
SET @medical_catalog_id = (SELECT `id` FROM `sys_menu` WHERE `path` = '/medical' AND `del_flag` = 0 ORDER BY `create_date` LIMIT 1);

INSERT INTO `sys_menu` (`id`, `pid`, `type`, `icon`, `component`, `title`, `name`, `path`, `auth_code`, `status`, `order`, `max_num_of_open_tab`, `create_date`, `update_date`, `del_flag`)
SELECT UUID(), @medical_catalog_id, 'menu', 'lucide:hospital', '/medical/department/list', 'medical.department.title', 'MedicalDepartment', '/medical/department', 'medical:department:list', 1, 1, -1, NOW(), NOW(), 0
WHERE NOT EXISTS (SELECT 1 FROM `sys_menu` WHERE `path` = '/medical/department' AND `del_flag` = 0);
SET @medical_department_menu_id = (SELECT `id` FROM `sys_menu` WHERE `path` = '/medical/department' AND `del_flag` = 0 ORDER BY `create_date` LIMIT 1);

INSERT INTO `sys_menu` (`id`, `pid`, `type`, `title`, `name`, `auth_code`, `status`, `order`, `max_num_of_open_tab`, `create_date`, `update_date`, `del_flag`)
SELECT UUID(), @medical_department_menu_id, 'button', 'medical.permission.departmentCreate', 'MedicalDepartmentCreate', 'medical:department:create', 1, 1, -1, NOW(), NOW(), 0
WHERE NOT EXISTS (SELECT 1 FROM `sys_menu` WHERE `auth_code` = 'medical:department:create' AND `del_flag` = 0);
INSERT INTO `sys_menu` (`id`, `pid`, `type`, `title`, `name`, `auth_code`, `status`, `order`, `max_num_of_open_tab`, `create_date`, `update_date`, `del_flag`)
SELECT UUID(), @medical_department_menu_id, 'button', 'medical.permission.departmentUpdate', 'MedicalDepartmentUpdate', 'medical:department:update', 1, 2, -1, NOW(), NOW(), 0
WHERE NOT EXISTS (SELECT 1 FROM `sys_menu` WHERE `auth_code` = 'medical:department:update' AND `del_flag` = 0);
INSERT INTO `sys_menu` (`id`, `pid`, `type`, `title`, `name`, `auth_code`, `status`, `order`, `max_num_of_open_tab`, `create_date`, `update_date`, `del_flag`)
SELECT UUID(), @medical_department_menu_id, 'button', 'medical.permission.departmentStatus', 'MedicalDepartmentStatus', 'medical:department:status', 1, 3, -1, NOW(), NOW(), 0
WHERE NOT EXISTS (SELECT 1 FROM `sys_menu` WHERE `auth_code` = 'medical:department:status' AND `del_flag` = 0);
INSERT INTO `sys_menu` (`id`, `pid`, `type`, `title`, `name`, `auth_code`, `status`, `order`, `max_num_of_open_tab`, `create_date`, `update_date`, `del_flag`)
SELECT UUID(), @medical_department_menu_id, 'button', 'medical.permission.departmentDelete', 'MedicalDepartmentDelete', 'medical:department:delete', 1, 4, -1, NOW(), NOW(), 0
WHERE NOT EXISTS (SELECT 1 FROM `sys_menu` WHERE `auth_code` = 'medical:department:delete' AND `del_flag` = 0);

INSERT INTO `sys_menu` (`id`, `pid`, `type`, `icon`, `component`, `title`, `name`, `path`, `auth_code`, `status`, `order`, `max_num_of_open_tab`, `create_date`, `update_date`, `del_flag`)
SELECT UUID(), @medical_catalog_id, 'menu', 'lucide:user-round', '/medical/doctor/list', 'medical.doctor.title', 'MedicalDoctor', '/medical/doctor', 'medical:doctor:list', 1, 2, -1, NOW(), NOW(), 0
WHERE NOT EXISTS (SELECT 1 FROM `sys_menu` WHERE `path` = '/medical/doctor' AND `del_flag` = 0);
SET @medical_doctor_menu_id = (SELECT `id` FROM `sys_menu` WHERE `path` = '/medical/doctor' AND `del_flag` = 0 ORDER BY `create_date` LIMIT 1);

INSERT INTO `sys_menu` (`id`, `pid`, `type`, `title`, `name`, `auth_code`, `status`, `order`, `max_num_of_open_tab`, `create_date`, `update_date`, `del_flag`)
SELECT UUID(), @medical_doctor_menu_id, 'button', 'medical.permission.doctorDetail', 'MedicalDoctorDetail', 'medical:doctor:detail', 1, 1, -1, NOW(), NOW(), 0
WHERE NOT EXISTS (SELECT 1 FROM `sys_menu` WHERE `auth_code` = 'medical:doctor:detail' AND `del_flag` = 0);
INSERT INTO `sys_menu` (`id`, `pid`, `type`, `title`, `name`, `auth_code`, `status`, `order`, `max_num_of_open_tab`, `create_date`, `update_date`, `del_flag`)
SELECT UUID(), @medical_doctor_menu_id, 'button', 'medical.permission.doctorCreate', 'MedicalDoctorCreate', 'medical:doctor:create', 1, 2, -1, NOW(), NOW(), 0
WHERE NOT EXISTS (SELECT 1 FROM `sys_menu` WHERE `auth_code` = 'medical:doctor:create' AND `del_flag` = 0);
INSERT INTO `sys_menu` (`id`, `pid`, `type`, `title`, `name`, `auth_code`, `status`, `order`, `max_num_of_open_tab`, `create_date`, `update_date`, `del_flag`)
SELECT UUID(), @medical_doctor_menu_id, 'button', 'medical.permission.doctorUpdate', 'MedicalDoctorUpdate', 'medical:doctor:update', 1, 3, -1, NOW(), NOW(), 0
WHERE NOT EXISTS (SELECT 1 FROM `sys_menu` WHERE `auth_code` = 'medical:doctor:update' AND `del_flag` = 0);
INSERT INTO `sys_menu` (`id`, `pid`, `type`, `title`, `name`, `auth_code`, `status`, `order`, `max_num_of_open_tab`, `create_date`, `update_date`, `del_flag`)
SELECT UUID(), @medical_doctor_menu_id, 'button', 'medical.permission.doctorStatus', 'MedicalDoctorStatus', 'medical:doctor:status', 1, 4, -1, NOW(), NOW(), 0
WHERE NOT EXISTS (SELECT 1 FROM `sys_menu` WHERE `auth_code` = 'medical:doctor:status' AND `del_flag` = 0);
INSERT INTO `sys_menu` (`id`, `pid`, `type`, `title`, `name`, `auth_code`, `status`, `order`, `max_num_of_open_tab`, `create_date`, `update_date`, `del_flag`)
SELECT UUID(), @medical_doctor_menu_id, 'button', 'medical.permission.doctorDelete', 'MedicalDoctorDelete', 'medical:doctor:delete', 1, 5, -1, NOW(), NOW(), 0
WHERE NOT EXISTS (SELECT 1 FROM `sys_menu` WHERE `auth_code` = 'medical:doctor:delete' AND `del_flag` = 0);
