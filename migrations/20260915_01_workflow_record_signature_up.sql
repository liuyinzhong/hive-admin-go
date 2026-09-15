-- 审批签署签名快照：wf_process_record 增加快照 URL 字段。
-- 上线顺序：先执行本脚本，再发布读写 signature 字段的后端版本。
-- 既有记录签名为 NULL，不回填：历史记录签署时没有快照机制，禁止伪造证据。

ALTER TABLE `wf_process_record`
  ADD COLUMN `signature` varchar(256) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT '签署人签名快照URL' AFTER `comment`;
