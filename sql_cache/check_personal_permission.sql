-- 个人权限功能 403 排查脚本（一次性使用）
-- 用法：把 @user_name 替换为当前登录账号的登录名后整段执行，按注释逐段看结果。

SET @user_name = '这里换成登录名';

-- ① 找到操作者 userId
SET @uid = (SELECT user_id FROM sys_user WHERE username = @user_name AND del_flag = 0 LIMIT 1);
SELECT @uid AS 操作者userId;

-- ② 按钮节点是否正确创建：应有一条 auth_code 完全等于 system:user:personalPermission
--    常见错误：拼写不一（personalpermission 大小写）、status=0（停用）、del_flag=1、title 随意没关系
SELECT id, title, auth_code, status, del_flag, pid
FROM sys_menu
WHERE auth_code LIKE 'system:user:personal%';

-- ③ 操作者的有效角色：应至少一条且 status=1
SELECT r.role_id, r.role_title, r.status
FROM sys_user_role ur
JOIN sys_role r ON r.role_id = ur.role_id AND r.del_flag = 0
WHERE ur.user_id = @uid AND ur.del_flag = 0;

-- ④ ②中的按钮节点是否被这些角色勾选（把上一步的 role_id 对照，或直接看总数）
SELECT rm.role_id, r.role_title
FROM sys_role_menu rm
JOIN sys_menu m ON m.id = rm.menu_id AND m.del_flag = 0
JOIN sys_role r ON r.role_id = rm.role_id AND r.del_flag = 0
WHERE m.auth_code = 'system:user:personalPermission' AND rm.del_flag = 0;

-- ⑤ 终判：模拟后端实时口径，有结果 = 权限已生效（刷新页面即可用）；无结果 = 未生效
SELECT DISTINCT m.auth_code
FROM sys_user_role ur
JOIN sys_role r ON r.role_id = ur.role_id AND r.status = 1 AND r.del_flag = 0
JOIN sys_role_menu rm ON rm.role_id = r.role_id AND rm.del_flag = 0
JOIN sys_menu m ON m.id = rm.menu_id AND m.status = 1 AND m.del_flag = 0
WHERE ur.user_id = @uid AND ur.del_flag = 0
  AND m.auth_code IN ('system:user:permission', 'system:user:personalPermission', 'system:menu:detail');
