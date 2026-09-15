-- 存量用户密码 bcrypt 化批量重置（配合 20260914_01_user_pwd_version_up.sql 使用）。
-- 背景：密码存储从明文切换为 bcrypt（见 business-docs/system/access-control.md SYS-ACL-040），
-- 不再清空重建数据库，改为把存量用户口令统一重置为已知值，用户登录后可通过修改密码接口自助换新。
--
-- 统一明文：123456（不满足 SYS-ACL-041 强度策略，仅为过渡值；SQL 直写不经过接口校验）
-- 范围：与盘点口径一致——未删除且非 superAdmin 的用户；superAdmin 口令已由
--       sql_cache/20260915_superadmin_bcrypt_password.sql 单独处理。
-- 副作用：pwd_version 递增使全部旧会话凭证立即失效，所有用户需重新登录（与 SYS-ACL-043 重置语义一致）。
-- 风险：所有存量用户共用同一已知口令，且登录尚无失败锁定/限流，执行后应尽快通知用户各自修改密码；
--       原明文口令被覆盖且哈希不可逆，本脚本没有回滚脚本，执行前请自行备份数据库。
-- 兼容性：升级前签发的旧 token 不携带版本号（解析为 0），与递增后的版本号不匹配，同样被强制下线。

UPDATE `sys_user`
SET `password`    = '$2a$10$DahHew2o08ntsD4npsynwea6j2tk1WnIHX1GWwCVAw.wAGDtq68M.',
    `pwd_version` = `pwd_version` + 1,
    `update_date` = NOW()
WHERE `del_flag` = 0
  AND `username` != 'superAdmin';
