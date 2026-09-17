-- 回滚用户个人权限表。回滚前需确认 sys_user_menu 已无数据或数据可丢弃：
-- 回滚后所有个人额外授权与个人禁止失效，用户生效权限回到纯角色口径。
DROP TABLE IF EXISTS `sys_user_menu`;
