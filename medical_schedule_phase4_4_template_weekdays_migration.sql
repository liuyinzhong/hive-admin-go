-- 医生出诊排班增量：一条周期排班模板支持多个适用星期。
-- 项目使用逻辑关联，不创建物理外键。
-- 本脚本可重复执行，不包含 DELETE、DROP 或 TRUNCATE。

CREATE TABLE IF NOT EXISTS `med_schedule_template_weekday` (
  `template_id` CHAR(36) NOT NULL COMMENT '周期排班模板ID',
  `weekday` TINYINT NOT NULL COMMENT '星期 1周一至7周日',
  `create_date` DATETIME NOT NULL COMMENT '创建时间',
  PRIMARY KEY (`template_id`, `weekday`),
  KEY `idx_med_schedule_template_weekday` (`weekday`, `template_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='周期排班模板适用星期';

-- 旧模型每条模板只有一个星期，先完整回填关系表。
INSERT IGNORE INTO `med_schedule_template_weekday` (`template_id`, `weekday`, `create_date`)
SELECT `template_id`, `weekday`, COALESCE(`create_date`, NOW())
FROM `med_schedule_template`
WHERE `weekday` BETWEEN 1 AND 7;

-- 合并由同一次星期多选误拆出的模板。除星期、主键和审计字段外，业务配置必须完全一致。
-- 先把所有重复模板的星期并入字典序最小的模板ID，再逻辑停用其余主表记录。
INSERT IGNORE INTO `med_schedule_template_weekday` (`template_id`, `weekday`, `create_date`)
SELECT canonical.`template_id`, duplicate_weekday.`weekday`, NOW()
FROM `med_schedule_template` canonical
JOIN `med_schedule_template` duplicate
  ON canonical.`template_id` < duplicate.`template_id`
  AND canonical.`template_name` = duplicate.`template_name`
  AND canonical.`doctor_id` = duplicate.`doctor_id`
  AND canonical.`department_id` = duplicate.`department_id`
  AND canonical.`registration_type` = duplicate.`registration_type`
  AND canonical.`start_time` = duplicate.`start_time`
  AND canonical.`end_time` = duplicate.`end_time`
  AND canonical.`default_slot_quota` = duplicate.`default_slot_quota`
  AND canonical.`slot_quota_config` <=> duplicate.`slot_quota_config`
  AND canonical.`total_quota` = duplicate.`total_quota`
  AND canonical.`effective_date` = duplicate.`effective_date`
  AND canonical.`expiry_date` <=> duplicate.`expiry_date`
  AND canonical.`status` = duplicate.`status`
  AND canonical.`remark` <=> duplicate.`remark`
  AND canonical.`del_flag` = 0
  AND duplicate.`del_flag` = 0
JOIN `med_schedule_template_weekday` duplicate_weekday
  ON duplicate_weekday.`template_id` = duplicate.`template_id`;

UPDATE `med_schedule_template` duplicate
JOIN `med_schedule_template` canonical
  ON canonical.`template_id` < duplicate.`template_id`
  AND canonical.`template_name` = duplicate.`template_name`
  AND canonical.`doctor_id` = duplicate.`doctor_id`
  AND canonical.`department_id` = duplicate.`department_id`
  AND canonical.`registration_type` = duplicate.`registration_type`
  AND canonical.`start_time` = duplicate.`start_time`
  AND canonical.`end_time` = duplicate.`end_time`
  AND canonical.`default_slot_quota` = duplicate.`default_slot_quota`
  AND canonical.`slot_quota_config` <=> duplicate.`slot_quota_config`
  AND canonical.`total_quota` = duplicate.`total_quota`
  AND canonical.`effective_date` = duplicate.`effective_date`
  AND canonical.`expiry_date` <=> duplicate.`expiry_date`
  AND canonical.`status` = duplicate.`status`
  AND canonical.`remark` <=> duplicate.`remark`
  AND canonical.`del_flag` = 0
  AND duplicate.`del_flag` = 0
SET duplicate.`status` = 0,
    duplicate.`del_flag` = 1,
    duplicate.`update_date` = NOW();

-- 保留旧字段作为兼容回退值，新代码的完整星期集合以关系表为准。
UPDATE `med_schedule_template` template
JOIN (
  SELECT `template_id`, MIN(`weekday`) AS `first_weekday`
  FROM `med_schedule_template_weekday`
  GROUP BY `template_id`
) weekday_set ON weekday_set.`template_id` = template.`template_id`
SET template.`weekday` = weekday_set.`first_weekday`
WHERE template.`del_flag` = 0;
