-- ----------------------------
-- Table structure for wf_process_definition
-- ----------------------------
CREATE TABLE IF NOT EXISTS `wf_process_definition` (
  `definition_id` char(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '流程定义id,UUID格式',
  `definition_key` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '流程标识,系统内唯一',
  `definition_name` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '流程名称',
  `category` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT '流程分类',
  `status` tinyint NULL DEFAULT 0 COMMENT '状态:0草稿 1已发布 2已停用',
  `version` int NULL DEFAULT 0 COMMENT '发布版本号',
  `flow_data` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL COMMENT 'LogicFlow画布JSON',
  `remark` varchar(256) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT '备注',
  `creator_id` char(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT '创建人id',
  `create_date` datetime NULL DEFAULT NULL COMMENT '创建日期',
  `update_date` datetime NULL DEFAULT NULL COMMENT '修改日期',
  `del_flag` tinyint NULL DEFAULT 0 COMMENT '逻辑删除 0:正常 1:删除',
  PRIMARY KEY (`definition_id`) USING BTREE,
  INDEX `idx_wf_process_definition_key` (`definition_key`) USING BTREE,
  INDEX `idx_wf_process_definition_status` (`status`) USING BTREE,
  INDEX `idx_wf_process_definition_category` (`category`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '流程定义表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Optional menu seed for workflow definition
-- ----------------------------
INSERT INTO `sys_menu` (
  `id`, `pid`, `type`, `icon`, `active_icon`, `keep_alive`, `hide_in_menu`, `hide_in_tab`, `hide_in_breadcrumb`,
  `hide_children_in_menu`, `badge`, `badge_type`, `badge_variants`, `active_path`, `auth_code`, `affix_tab`,
  `component`, `title`, `name`, `path`, `status`, `link`, `iframe_src`, `order`, `max_num_of_open_tab`,
  `affix_tab_order`, `no_basic_layout`, `open_in_new_window`, `dom_cached`, `query`, `menu_visible_with_forbidden`,
  `creator_id`, `creator_name`, `create_date`, `update_date`, `del_flag`
)
SELECT
  'd42e53e7-87bd-4dbd-9d51-f0fd9e6c4201', NULL, 'catalog', 'lucide:workflow', NULL, 0, 0, 0, 0,
  0, NULL, NULL, NULL, NULL, 'workflow', 0,
  NULL, '流程管理', 'Workflow', '/workflow', 1, NULL, NULL, 50, -1,
  0, 0, 0, 0, NULL, 0,
  NULL, NULL, NOW(), NOW(), 0
WHERE NOT EXISTS (
  SELECT 1 FROM `sys_menu` WHERE `id` = 'd42e53e7-87bd-4dbd-9d51-f0fd9e6c4201'
);

INSERT INTO `sys_menu` (
  `id`, `pid`, `type`, `icon`, `active_icon`, `keep_alive`, `hide_in_menu`, `hide_in_tab`, `hide_in_breadcrumb`,
  `hide_children_in_menu`, `badge`, `badge_type`, `badge_variants`, `active_path`, `auth_code`, `affix_tab`,
  `component`, `title`, `name`, `path`, `status`, `link`, `iframe_src`, `order`, `max_num_of_open_tab`,
  `affix_tab_order`, `no_basic_layout`, `open_in_new_window`, `dom_cached`, `query`, `menu_visible_with_forbidden`,
  `creator_id`, `creator_name`, `create_date`, `update_date`, `del_flag`
)
SELECT
  '2d803c6e-c931-47b6-9331-a2efdb23f2a1', 'd42e53e7-87bd-4dbd-9d51-f0fd9e6c4201', 'menu', 'lucide:git-branch-plus', NULL, 0, 0, 0, 0,
  0, NULL, NULL, NULL, NULL, 'workflow:definition:list', 0,
  '/workflow/definition/list', '流程定义', 'WorkflowDefinition', '/workflow/definition/list', 1, NULL, NULL, 1, -1,
  0, 0, 0, 0, NULL, 0,
  NULL, NULL, NOW(), NOW(), 0
WHERE NOT EXISTS (
  SELECT 1 FROM `sys_menu` WHERE `id` = '2d803c6e-c931-47b6-9331-a2efdb23f2a1'
);
