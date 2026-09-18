-- dev_change_history 增加变更明细列
-- 存 JSON 数组,元素结构 {fieldKey,fieldLabel,dictType,oldValue,newValue};
-- 存量记录该列为 NULL,前端按无明细降级展示,无需数据回填。
ALTER TABLE `dev_change_history`
    ADD COLUMN `change_items` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL COMMENT '变更明细JSON数组,元素结构 {fieldKey,fieldLabel,dictType,oldValue,newValue}' AFTER `extend_json`;
