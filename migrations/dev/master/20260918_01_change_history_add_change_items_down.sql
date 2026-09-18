-- 回滚:删除变更明细列,存量明细数据随列删除,不影响变更记录其它字段
ALTER TABLE `dev_change_history` DROP COLUMN `change_items`;
