-- 审批签署签名快照迁移回滚。执行前应确认应用已回退到不读写 signature 字段的版本。
-- 快照文件（static/uploads/workflow-sign/）需另行手动清理，本脚本只回滚表结构。

ALTER TABLE `wf_process_record`
  DROP COLUMN `signature`;
