-- 回滚令牌黑名单表。回滚后登出恢复为无服务端失效，已撤销凭证在自然过期前重新可用，
-- 需回退对应版本的服务端代码后执行本脚本。
DROP TABLE IF EXISTS `sys_token_blacklist`;
