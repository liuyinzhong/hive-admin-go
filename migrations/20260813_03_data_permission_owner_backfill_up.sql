-- 数据权限归属修复：为历史自动生成排班继承其周期模板创建人。
-- 仅处理“自动生成批次创建人为空、实际排班创建人为空、模板创建人明确”的记录；
-- 无法可靠推断归属的历史数据保持为空，并仅对全部数据权限可见。

UPDATE `med_schedule` AS `schedule`
INNER JOIN `med_schedule_generation_batch` AS `batch`
  ON `batch`.`batch_id` = `schedule`.`generation_batch_id`
INNER JOIN `med_schedule_template` AS `template`
  ON `template`.`template_id` = `schedule`.`template_id`
SET
  `schedule`.`creator_id` = `template`.`creator_id`,
  `schedule`.`updater_id` = COALESCE(`schedule`.`updater_id`, `template`.`creator_id`)
WHERE `batch`.`creator_id` IS NULL
  AND `schedule`.`creator_id` IS NULL
  AND `template`.`creator_id` IS NOT NULL;
