-- 一次性操作 SQL：清空重建数据库后，把超级管理员种子账号的明文密码更新为 bcrypt 哈希。
-- 背景：代码已全面切换为 bcrypt 比对（见 business-docs/system/access-control.md SYS-ACL-040），
-- hive.sql 备份中的明文种子密码无法登录。
-- 明文密码：Hive@2026 （满足强度策略：长度≥8，含大写、小写、数字、特殊字符）
-- 执行后请用 superAdmin / Hive@2026 登录，并尽快通过页面修改为已知口令。
-- 不使用时可直接删除本文件。

UPDATE `sys_user`
SET `password` = '$2a$10$YgiD4PTUvDA1e5LV5BPnqOylAzLHQkbjn0U.lSk/BMmnm44TZnIhC'
WHERE `username` = 'superAdmin' AND `del_flag` = 0;
