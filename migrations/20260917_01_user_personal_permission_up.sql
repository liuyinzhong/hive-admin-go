-- 用户个人权限：直接作用于单个用户的菜单授权调整表。
-- 语义模型见 business-docs/system/adr/0006-user-personal-permission-tri-state.md：
-- 用户生效权限 = 有效角色授权 ∪ grant(个人额外授权) − deny(个人禁止)，禁止绝对优先。
-- 仅影响功能权限（菜单与按钮），数据范围仍完全由角色决定；系统内置超级用户不读本表。
-- 删除用户时随用户关联一并物理清理；删除菜单时随角色菜单关联一并物理清理。
-- 本脚本为全新空表，无存量数据；collation 与业务关联表（sys_role_dept 等）保持 utf8mb4_unicode_ci，
-- 避免与 sys_user JOIN 时排序规则冲突。
-- 若此前已按旧版（utf8mb4_general_ci）建过表：先执行 DROP TABLE sys_user_menu 再执行本脚本重建。
CREATE TABLE `sys_user_menu` (
    `id`          char(36)     CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '主键UUID',
    `user_id`     char(36)     CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '系统用户ID',
    `menu_id`     char(36)     CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '菜单节点ID（sys_menu.id，含按钮节点）',
    `grant_type`  varchar(16)  CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '个人权限类型 grant=额外授权 deny=禁止',
    `create_date` datetime     NULL COMMENT '创建时间',
    `update_date` datetime     NULL COMMENT '更新时间',
    `del_flag`    int          NOT NULL DEFAULT 0 COMMENT '删除标记 0=正常 1=已删除',
    PRIMARY KEY (`id`),
    KEY `idx_user_type` (`user_id`, `grant_type`),
    KEY `idx_menu` (`menu_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户个人权限（额外授权与禁止）';
