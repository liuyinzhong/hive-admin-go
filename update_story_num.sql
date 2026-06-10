-- 修改 dev_change_history 表的 business_id 字段长度（支持 UUID 格式）
ALTER TABLE dev_change_history MODIFY COLUMN business_id char(36) NOT NULL COMMENT '外键业务id,逻辑外键';

-- 修改 dev_story 表的 story_num 字段为自增长
ALTER TABLE dev_story MODIFY COLUMN story_num int NOT NULL AUTO_INCREMENT COMMENT '需求编号';

-- 修改 dev_task 表的 task_num 字段为自增长
ALTER TABLE dev_task MODIFY COLUMN task_num int NOT NULL AUTO_INCREMENT COMMENT '任务编号';

-- 修改 dev_bug 表的 bug_num 字段为自增长
ALTER TABLE dev_bug MODIFY COLUMN bug_num int NOT NULL AUTO_INCREMENT COMMENT 'bug编号';
