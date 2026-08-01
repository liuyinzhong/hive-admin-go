CREATE TABLE IF NOT EXISTS `sys_menu_message` (
  `id` CHAR(36) NOT NULL COMMENT '消息记录ID',
  `user_id` CHAR(36) NOT NULL COMMENT '接收用户ID',
  `menu_id` CHAR(36) NOT NULL COMMENT '目标菜单ID',
  `title` VARCHAR(128) NOT NULL COMMENT '消息标题',
  `content` VARCHAR(512) NOT NULL COMMENT '消息内容',
  `read_at` DATETIME NULL COMMENT '已读时间，NULL表示未读',
  `create_date` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_sys_menu_message_user_unread` (`user_id`, `read_at`),
  KEY `idx_sys_menu_message_user_menu_unread` (`user_id`, `menu_id`, `read_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户菜单消息表';
