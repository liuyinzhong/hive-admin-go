-- 回滚用户密码版本号列。回滚后已签发携带版本号的 token 与 JWT claims 结构不匹配的校验随之失效，
-- 需回退对应版本的服务端代码后执行本脚本。
ALTER TABLE `sys_user`
    DROP COLUMN `pwd_version`;
