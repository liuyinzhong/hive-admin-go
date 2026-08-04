-- 打印模板主表。
-- 只保存当前草稿和当前已发布内容，不保存历史版本；执行前由部署人员按项目迁移流程确认数据库备份与回滚窗口。
CREATE TABLE print_template (
    template_id      CHAR(36)     NOT NULL,
    document_type    VARCHAR(64)  NOT NULL,
    template_name    VARCHAR(128) NOT NULL,
    draft_layout     LONGTEXT     NOT NULL,
    published_layout LONGTEXT     NULL,
    status           VARCHAR(16)  NOT NULL,
    row_version      INT          NOT NULL DEFAULT 1,
    creator_id       CHAR(36)     NULL,
    updater_id       CHAR(36)     NULL,
    create_date      DATETIME     NULL,
    update_date      DATETIME     NULL,
    PRIMARY KEY (template_id),
    UNIQUE KEY uk_print_template_document_type (document_type),
    KEY idx_print_template_status (status),
    KEY idx_print_template_update_date (update_date)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci;

