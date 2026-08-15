-- 回滚历史自动排班归属。
-- 该条件也会覆盖升级后新生成且仍由同一模板创建人持有的自动排班，
-- 因此仅应与旧版本应用回滚同时执行，并应先备份相关排班记录。

UPDATE `med_schedule` AS `schedule`
INNER JOIN `med_schedule_generation_batch` AS `batch`
  ON `batch`.`batch_id` = `schedule`.`generation_batch_id`
INNER JOIN `med_schedule_template` AS `template`
  ON `template`.`template_id` = `schedule`.`template_id`
SET
  `schedule`.`updater_id` = CASE
    WHEN `schedule`.`updater_id` = `template`.`creator_id` THEN NULL
    ELSE `schedule`.`updater_id`
  END,
  `schedule`.`creator_id` = NULL
WHERE `batch`.`creator_id` IS NULL
  AND `template`.`creator_id` IS NOT NULL
  AND `schedule`.`creator_id` = `template`.`creator_id`;
