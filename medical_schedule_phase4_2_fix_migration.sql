-- 医生出诊排班修正：仅系统管理员可见和操作排班菜单。
-- 系统管理员（sys_user.is_sys = 1）读取全部菜单，无需角色菜单授权。
-- 本脚本不包含 DELETE、DROP 或 TRUNCATE，可重复执行。

UPDATE `sys_role_menu` role_menu
JOIN `sys_menu` menu ON menu.`id` = role_menu.`menu_id` AND menu.`del_flag` = 0
SET role_menu.`del_flag` = 1, role_menu.`update_date` = NOW()
WHERE role_menu.`del_flag` = 0
  AND (
    menu.`path` = '/medical/schedule'
    OR menu.`path` LIKE '/medical/schedule/%'
    OR menu.`auth_code` LIKE 'medical:schedule:%'
    OR menu.`auth_code` LIKE 'medical:schedule-template:%'
    OR menu.`auth_code` LIKE 'medical:schedule-task:%'
  );
