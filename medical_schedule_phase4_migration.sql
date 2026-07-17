-- 医生管理第四阶段：周期排班模板、生成批次、实际排班及排班状态字典。
-- 项目使用逻辑关联，不创建物理外键。
-- 本脚本只包含创建和插入操作，可重复执行；不包含 DELETE、DROP 或 TRUNCATE。

CREATE TABLE IF NOT EXISTS `med_schedule_template` (
  `template_id` CHAR(36) NOT NULL COMMENT '周期排班模板ID',
  `template_name` VARCHAR(64) NOT NULL COMMENT '模板名称',
  `doctor_id` CHAR(36) NOT NULL COMMENT '医生ID',
  `department_id` CHAR(36) NOT NULL COMMENT '出诊科室ID',
  `registration_type` VARCHAR(36) NOT NULL COMMENT '挂号类型，字典MED_REGISTRATION_TYPE',
  `weekday` TINYINT NOT NULL COMMENT '星期 1周一至7周日',
  `start_time` TIME NOT NULL COMMENT '开始时间',
  `end_time` TIME NOT NULL COMMENT '结束时间',
  `total_quota` INT NOT NULL COMMENT '模板号源总量',
  `effective_date` DATE NOT NULL COMMENT '模板生效日期（含）',
  `expiry_date` DATE NULL COMMENT '模板失效日期（含），NULL表示长期有效',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 0停用 1启用',
  `remark` VARCHAR(512) NULL COMMENT '备注',
  `creator_id` CHAR(36) NULL COMMENT '创建人ID',
  `updater_id` CHAR(36) NULL COMMENT '更新人ID',
  `create_date` DATETIME NOT NULL COMMENT '创建时间',
  `update_date` DATETIME NOT NULL COMMENT '更新时间',
  `del_flag` TINYINT NOT NULL DEFAULT 0 COMMENT '删除标记 0正常 1删除',
  PRIMARY KEY (`template_id`),
  KEY `idx_med_schedule_template_doctor` (`doctor_id`, `weekday`, `status`, `del_flag`, `effective_date`, `expiry_date`),
  KEY `idx_med_schedule_template_department` (`department_id`, `status`, `del_flag`),
  KEY `idx_med_schedule_template_period` (`effective_date`, `expiry_date`, `status`, `del_flag`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='周期排班模板';

CREATE TABLE IF NOT EXISTS `med_schedule_generation_batch` (
  `batch_id` CHAR(36) NOT NULL COMMENT '排班生成批次ID',
  `idempotency_key` VARCHAR(64) NOT NULL COMMENT '客户端幂等键',
  `request_hash` CHAR(64) NOT NULL COMMENT '规范化请求SHA256',
  `template_ids` LONGTEXT NOT NULL COMMENT '模板ID列表JSON',
  `start_date` DATE NOT NULL COMMENT '生成开始日期（含）',
  `end_date` DATE NOT NULL COMMENT '生成结束日期（含）',
  `status` TINYINT NOT NULL DEFAULT 0 COMMENT '状态 0处理中 1完成',
  `generated_count` INT NOT NULL DEFAULT 0 COMMENT '生成数量',
  `skipped_count` INT NOT NULL DEFAULT 0 COMMENT '因幂等已存在而跳过数量',
  `creator_id` CHAR(36) NULL COMMENT '创建人ID',
  `create_date` DATETIME NOT NULL COMMENT '创建时间',
  `update_date` DATETIME NOT NULL COMMENT '更新时间',
  PRIMARY KEY (`batch_id`),
  UNIQUE KEY `uk_med_schedule_batch_idempotency` (`idempotency_key`),
  KEY `idx_med_schedule_batch_date` (`start_date`, `end_date`, `create_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='排班批量生成幂等批次';

CREATE TABLE IF NOT EXISTS `med_schedule` (
  `schedule_id` CHAR(36) NOT NULL COMMENT '实际排班ID',
  `template_id` CHAR(36) NULL COMMENT '来源周期模板ID，手工排班为空',
  `generation_batch_id` CHAR(36) NULL COMMENT '来源生成批次ID',
  `doctor_id` CHAR(36) NOT NULL COMMENT '医生ID',
  `department_id` CHAR(36) NOT NULL COMMENT '出诊科室ID',
  `registration_type` VARCHAR(36) NOT NULL COMMENT '挂号类型，字典MED_REGISTRATION_TYPE',
  `schedule_date` DATE NOT NULL COMMENT '出诊日期',
  `start_time` TIME NOT NULL COMMENT '开始时间',
  `end_time` TIME NOT NULL COMMENT '结束时间',
  `fee_rule_id` CHAR(36) NOT NULL COMMENT '挂号费规则ID快照来源',
  `fee_rule_version` INT NOT NULL COMMENT '挂号费规则版本快照',
  `fee_amount` DECIMAL(10,2) NOT NULL COMMENT '挂号费金额快照（元）',
  `total_quota` INT NOT NULL COMMENT '号源总量',
  `booked_quota` INT NOT NULL DEFAULT 0 COMMENT '已预约号源数',
  `status` TINYINT NOT NULL DEFAULT 0 COMMENT '状态 0草稿 1已发布 2停诊 3结束',
  `stop_reason` VARCHAR(512) NULL COMMENT '停诊原因',
  `published_at` DATETIME NULL COMMENT '发布时间',
  `stopped_at` DATETIME NULL COMMENT '停诊时间',
  `finished_at` DATETIME NULL COMMENT '结束时间',
  `remark` VARCHAR(512) NULL COMMENT '备注',
  `creator_id` CHAR(36) NULL COMMENT '创建人ID',
  `updater_id` CHAR(36) NULL COMMENT '更新人ID',
  `create_date` DATETIME NOT NULL COMMENT '创建时间',
  `update_date` DATETIME NOT NULL COMMENT '更新时间',
  `del_flag` TINYINT NOT NULL DEFAULT 0 COMMENT '删除标记 0正常 1删除',
  PRIMARY KEY (`schedule_id`),
  UNIQUE KEY `uk_med_schedule_template_date` (`template_id`, `schedule_date`),
  KEY `idx_med_schedule_doctor_conflict` (`doctor_id`, `schedule_date`, `status`, `del_flag`, `start_time`, `end_time`),
  KEY `idx_med_schedule_department_date` (`department_id`, `schedule_date`, `status`, `del_flag`),
  KEY `idx_med_schedule_batch` (`generation_batch_id`),
  KEY `idx_med_schedule_fee_rule` (`fee_rule_id`, `fee_rule_version`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='实际排班及费用号源快照';

-- 排班状态字典仅用于展示，状态迁移仍由后端固定规则控制。
INSERT INTO `sys_dict` (`id`, `pid`, `type`, `label`, `value`, `remark`, `create_date`, `update_date`, `del_flag`, `status`)
SELECT UUID(), NULL, 'MED_SCHEDULE_STATUS', '排班状态', NULL, '实际排班状态展示字典', NOW(), NOW(), 0, 1
WHERE NOT EXISTS (SELECT 1 FROM `sys_dict` WHERE `type` = 'MED_SCHEDULE_STATUS' AND `pid` IS NULL AND `del_flag` = 0);

SET @med_schedule_status_dict_id = (
  SELECT `id` FROM `sys_dict`
  WHERE `type` = 'MED_SCHEDULE_STATUS' AND `pid` IS NULL AND `del_flag` = 0
  ORDER BY `create_date` LIMIT 1
);

INSERT INTO `sys_dict` (`id`, `pid`, `type`, `label`, `value`, `remark`, `create_date`, `update_date`, `del_flag`, `status`)
SELECT UUID(), @med_schedule_status_dict_id, 'MED_SCHEDULE_STATUS', '草稿', '0', NULL, NOW(), NOW(), 0, 1
WHERE NOT EXISTS (SELECT 1 FROM `sys_dict` WHERE `type` = 'MED_SCHEDULE_STATUS' AND `value` = '0' AND `del_flag` = 0);

INSERT INTO `sys_dict` (`id`, `pid`, `type`, `label`, `value`, `remark`, `create_date`, `update_date`, `del_flag`, `status`)
SELECT UUID(), @med_schedule_status_dict_id, 'MED_SCHEDULE_STATUS', '已发布', '1', NULL, NOW(), NOW(), 0, 1
WHERE NOT EXISTS (SELECT 1 FROM `sys_dict` WHERE `type` = 'MED_SCHEDULE_STATUS' AND `value` = '1' AND `del_flag` = 0);

INSERT INTO `sys_dict` (`id`, `pid`, `type`, `label`, `value`, `remark`, `create_date`, `update_date`, `del_flag`, `status`)
SELECT UUID(), @med_schedule_status_dict_id, 'MED_SCHEDULE_STATUS', '停诊', '2', NULL, NOW(), NOW(), 0, 1
WHERE NOT EXISTS (SELECT 1 FROM `sys_dict` WHERE `type` = 'MED_SCHEDULE_STATUS' AND `value` = '2' AND `del_flag` = 0);

INSERT INTO `sys_dict` (`id`, `pid`, `type`, `label`, `value`, `remark`, `create_date`, `update_date`, `del_flag`, `status`)
SELECT UUID(), @med_schedule_status_dict_id, 'MED_SCHEDULE_STATUS', '结束', '3', NULL, NOW(), NOW(), 0, 1
WHERE NOT EXISTS (SELECT 1 FROM `sys_dict` WHERE `type` = 'MED_SCHEDULE_STATUS' AND `value` = '3' AND `del_flag` = 0);
