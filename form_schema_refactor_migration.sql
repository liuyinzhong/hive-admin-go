-- 独立表单 Schema 与流程关联重构。
-- 当前阶段不做 Schema 版本；流程实例继续保存发起时的 form_snapshot。
-- 项目沿用逻辑关联，不创建物理外键。

CREATE TABLE `sys_form_schema` (
  `form_schema_id` CHAR(36) NOT NULL COMMENT '表单Schema ID',
  `schema_key` VARCHAR(128) NOT NULL COMMENT 'Schema唯一标识',
  `schema_name` VARCHAR(128) NOT NULL COMMENT 'Schema名称',
  `category` VARCHAR(64) NULL COMMENT '分类',
  `schema_json` LONGTEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '可持久化Vben Form Schema JSON',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 0禁用 1启用',
  `remark` VARCHAR(256) NULL COMMENT '备注',
  `creator_id` CHAR(36) NULL COMMENT '创建人ID',
  `create_date` DATETIME NULL COMMENT '创建时间',
  `update_date` DATETIME NULL COMMENT '更新时间',
  `del_flag` TINYINT NOT NULL DEFAULT 0 COMMENT '删除标记 0正常 1删除',
  PRIMARY KEY (`form_schema_id`),
  UNIQUE KEY `uk_sys_form_schema_key` (`schema_key`),
  KEY `idx_sys_form_schema_status` (`status`, `del_flag`),
  KEY `idx_sys_form_schema_category` (`category`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='独立表单Schema';

ALTER TABLE `wf_process_definition`
  ADD COLUMN `form_schema_id` CHAR(36) NULL COMMENT '关联表单Schema ID' AFTER `flow_data`,
  ADD KEY `idx_wf_process_definition_form_schema_id` (`form_schema_id`),
  DROP COLUMN `form_schema`;

-- 创建独立表单管理菜单；不自动修改任何角色权限。
INSERT INTO `sys_menu` (
  `id`, `pid`, `type`, `icon`, `component`, `title`, `name`, `path`, `status`,
  `order`, `max_num_of_open_tab`, `create_date`, `update_date`, `del_flag`
)
SELECT
  UUID(), NULL, 'catalog', 'lucide:layout-template', 'BasicLayout', 'form.title',
  'FormManagement', '/form', 1, 45, -1, NOW(), NOW(), 0
WHERE NOT EXISTS (
  SELECT 1 FROM `sys_menu` WHERE `path` = '/form' AND `del_flag` = 0
);

SET @form_catalog_id = (
  SELECT `id` FROM `sys_menu`
  WHERE `path` = '/form' AND `del_flag` = 0
  ORDER BY `create_date` ASC
  LIMIT 1
);

INSERT INTO `sys_menu` (
  `id`, `pid`, `type`, `icon`, `component`, `title`, `name`, `path`, `status`,
  `order`, `max_num_of_open_tab`, `create_date`, `update_date`, `del_flag`
)
SELECT
  UUID(), @form_catalog_id, 'menu', 'lucide:list-tree', '/form/schema/list',
  'form.title', 'FormSchemaList', '/form/schema/list', 1, 1, -1, NOW(), NOW(), 0
WHERE NOT EXISTS (
  SELECT 1 FROM `sys_menu`
  WHERE `path` = '/form/schema/list' AND `del_flag` = 0
);

-- 回滚前需要确认流程定义不再引用独立表单。
-- DELETE FROM `sys_role_menu` WHERE `menu_id` IN (SELECT `id` FROM `sys_menu` WHERE `path` IN ('/form', '/form/schema/list'));
-- DELETE FROM `sys_menu` WHERE `path` IN ('/form/schema/list', '/form');
-- ALTER TABLE `wf_process_definition` ADD COLUMN `form_schema` LONGTEXT NULL AFTER `flow_data`;
-- ALTER TABLE `wf_process_definition` DROP INDEX `idx_wf_process_definition_form_schema_id`, DROP COLUMN `form_schema_id`;
-- DROP TABLE `sys_form_schema`;
