-- 用户密码存储加固：sys_user 增加密码版本号。
-- 口令每次修改或重置时递增；认证中间件比对 JWT 中的版本号与当前值不一致时返回 401，
-- 使该用户全部旧会话凭证立即失效。
-- 注意：升级前签发的旧 token 不携带版本号字段，解析为 0，与列默认值 0 一致；
-- 配合清空重建后全部会话重新登录，之后任意一次口令变更（版本号 ≥1）即可使存量 0 版本凭证全部失效。
ALTER TABLE `sys_user`
    ADD COLUMN `pwd_version` int NOT NULL DEFAULT 0 COMMENT '密码版本号，口令修改或重置时递增，旧世代会话凭证随之失效' AFTER `password`;
