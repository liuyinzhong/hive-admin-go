-- 数据权限：角色数据范围与自定义部门。
-- 上线顺序：先执行本脚本，再发布依赖 data_scope 字段的后端版本。

ALTER TABLE `sys_role`
  ADD COLUMN `data_scope` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'self' COMMENT '数据范围:all,customDepartment,department,departmentAndChildren,self,none' AFTER `remark`;

-- 保持存量角色升级前的可见范围；上线后由管理员逐角色收紧。
UPDATE `sys_role`
SET `data_scope` = 'all'
WHERE `del_flag` = 0;

CREATE TABLE `sys_role_dept` (
  `id` char(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '主键ID,UUID格式',
  `role_id` char(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '角色ID',
  `dept_id` char(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '自定义数据范围部门ID',
  `create_date` datetime NOT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `uk_sys_role_dept_role_dept` (`role_id`, `dept_id`) USING BTREE,
  KEY `idx_sys_role_dept_dept` (`dept_id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '角色自定义数据范围部门关联表' ROW_FORMAT = Dynamic;
