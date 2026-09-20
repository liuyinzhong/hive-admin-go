-- 个人权限 bug 一次性修复脚本：
-- 此前保存「禁止」时，权限树（includeIndeterminate）会把勾选页面的半选父目录一并写入 deny，
-- 合并时父目录被整组逐出生效菜单，导致同目录下未被禁止的页面也一起消失。
-- 代码已在保存侧剔除 deny 集合中的目录节点；本脚本清理修复前已写入的 deny 目录记录。
-- 执行前可先 SELECT 确认影响范围。
SELECT um.id, um.user_id, um.menu_id, m.title
FROM sys_user_menu um
JOIN sys_menu m ON m.id = um.menu_id AND m.type = 'catalog'
WHERE um.grant_type = 'deny' AND um.del_flag = 0;

DELETE um FROM sys_user_menu um
JOIN sys_menu m ON m.id = um.menu_id AND m.type = 'catalog'
WHERE um.grant_type = 'deny' AND um.del_flag = 0;
