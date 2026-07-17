-- 医生出诊排班修正：自动任务增加“执行中”状态，用唯一任务键原子抢占执行权。
-- 本脚本不包含 DELETE、DROP 或 TRUNCATE，可重复执行。

ALTER TABLE `med_schedule_auto_task`
  MODIFY COLUMN `status` TINYINT NOT NULL COMMENT '状态 0成功 1部分成功 2失败 3执行中';
