# 03: 任务写入点接入变更明细

**What to build:** 任务的整体修改、局部字段修改（userId/taskType/startDate/endDate）、流转在业务事务内先读旧值、计算字段级 diff，写入变更明细。负责人变化翻译为人名，日期按项目时间格式化。

**Blocked by:** 01 后端变更明细基础设施

**Status:** ready-for-agent

- [ ] 整体修改生成变化字段的明细（含负责人、关联需求标题、版本名、模块名翻译）
- [ ] 局部修改生成单字段明细
- [ ] 流转生成状态 from→to 明细，正文行为不变
- [ ] gofmt / go vet / go build 通过
