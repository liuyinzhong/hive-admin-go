-- 外部页面管理：复用 sys_menu，不自动创建任何外部页面数据。
-- 执行前请先备份 sys_menu；本文件仅生成迁移内容，不由应用自动执行。

SET @add_ignore_access_column := IF(
  EXISTS (
    SELECT 1
    FROM `information_schema`.`columns`
    WHERE `table_schema` = DATABASE()
      AND `table_name` = 'sys_menu'
      AND `column_name` = 'ignore_access'
  ),
  'SELECT 1',
  'ALTER TABLE `sys_menu` ADD COLUMN `ignore_access` TINYINT NOT NULL DEFAULT 0 COMMENT ''是否忽略访问鉴权：0否，1是'' AFTER `no_basic_layout`'
);
PREPARE add_ignore_access_column_stmt FROM @add_ignore_access_column;
EXECUTE add_ignore_access_column_stmt;
DEALLOCATE PREPARE add_ignore_access_column_stmt;

START TRANSACTION;

SET @external_page_permission_parent := (
  SELECT `pid`
  FROM `sys_menu`
  WHERE FIND_IN_SET('system:menu:list', REPLACE(COALESCE(`auth_code`, ''), ' ', '')) > 0
    AND `type` = 'button'
    AND `del_flag` = 0
  LIMIT 1
);

INSERT INTO `sys_menu` (`id`, `pid`, `type`, `title`, `auth_code`, `status`, `order`, `create_date`, `update_date`, `del_flag`)
SELECT UUID(), @external_page_permission_parent, 'button', 'system.permission.externalPageList', 'system:externalPage:list', 1, 7, NOW(), NOW(), 0
WHERE @external_page_permission_parent IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM `sys_menu` WHERE FIND_IN_SET('system:externalPage:list', REPLACE(COALESCE(`auth_code`, ''), ' ', '')) > 0 AND `del_flag` = 0);

INSERT INTO `sys_menu` (`id`, `pid`, `type`, `title`, `auth_code`, `status`, `order`, `create_date`, `update_date`, `del_flag`)
SELECT UUID(), @external_page_permission_parent, 'button', 'system.permission.externalPageCreate', 'system:externalPage:create', 1, 8, NOW(), NOW(), 0
WHERE @external_page_permission_parent IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM `sys_menu` WHERE FIND_IN_SET('system:externalPage:create', REPLACE(COALESCE(`auth_code`, ''), ' ', '')) > 0 AND `del_flag` = 0);

INSERT INTO `sys_menu` (`id`, `pid`, `type`, `title`, `auth_code`, `status`, `order`, `create_date`, `update_date`, `del_flag`)
SELECT UUID(), @external_page_permission_parent, 'button', 'system.permission.externalPageDetail', 'system:externalPage:detail', 1, 9, NOW(), NOW(), 0
WHERE @external_page_permission_parent IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM `sys_menu` WHERE FIND_IN_SET('system:externalPage:detail', REPLACE(COALESCE(`auth_code`, ''), ' ', '')) > 0 AND `del_flag` = 0);

INSERT INTO `sys_menu` (`id`, `pid`, `type`, `title`, `auth_code`, `status`, `order`, `create_date`, `update_date`, `del_flag`)
SELECT UUID(), @external_page_permission_parent, 'button', 'system.permission.externalPageUpdate', 'system:externalPage:update', 1, 10, NOW(), NOW(), 0
WHERE @external_page_permission_parent IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM `sys_menu` WHERE FIND_IN_SET('system:externalPage:update', REPLACE(COALESCE(`auth_code`, ''), ' ', '')) > 0 AND `del_flag` = 0);

INSERT INTO `sys_menu` (`id`, `pid`, `type`, `title`, `auth_code`, `status`, `order`, `create_date`, `update_date`, `del_flag`)
SELECT UUID(), @external_page_permission_parent, 'button', 'system.permission.externalPageStatus', 'system:externalPage:status', 1, 11, NOW(), NOW(), 0
WHERE @external_page_permission_parent IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM `sys_menu` WHERE FIND_IN_SET('system:externalPage:status', REPLACE(COALESCE(`auth_code`, ''), ' ', '')) > 0 AND `del_flag` = 0);

INSERT INTO `sys_menu` (`id`, `pid`, `type`, `title`, `auth_code`, `status`, `order`, `create_date`, `update_date`, `del_flag`)
SELECT UUID(), @external_page_permission_parent, 'button', 'system.permission.externalPageDelete', 'system:externalPage:delete', 1, 12, NOW(), NOW(), 0
WHERE @external_page_permission_parent IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM `sys_menu` WHERE FIND_IN_SET('system:externalPage:delete', REPLACE(COALESCE(`auth_code`, ''), ' ', '')) > 0 AND `del_flag` = 0);

COMMIT;

-- 回滚参考（需人工确认无外部页面数据后执行）：
-- UPDATE `sys_menu` SET `del_flag` = 1, `update_date` = NOW()
-- WHERE `auth_code` IN (
--   'system:externalPage:list', 'system:externalPage:create',
--   'system:externalPage:detail', 'system:externalPage:update',
--   'system:externalPage:status', 'system:externalPage:delete'
-- ) AND `del_flag` = 0;
-- ALTER TABLE `sys_menu` DROP COLUMN `ignore_access`;
