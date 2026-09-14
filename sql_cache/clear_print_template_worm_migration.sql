-- 打印模板存量数据清空（一次性）
-- 背景：版式协议从自研 PrintLayout 切换为 worm-vue3-print TemplateData，两者 JSON 结构互不兼容。
-- 范围：仅 print_template 表；模板是全局主数据，不含业务单据数据。
-- 执行时机：部署新版后端前或部署后立即执行；执行后需在打印管理里重新创建并发布模板。
-- 回滚方案：无需回滚（旧版式数据对新版后端无意义，保留反而会让详情接口返回旧结构）。

DELETE FROM print_template;
