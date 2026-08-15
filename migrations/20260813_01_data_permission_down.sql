-- 数据权限迁移回滚。执行前应确认应用已回退到不读取 data_scope 的版本。

DROP TABLE IF EXISTS `sys_role_dept`;

ALTER TABLE `sys_role`
  DROP COLUMN `data_scope`;
