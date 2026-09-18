# 04: 缺陷写入点接入变更明细

**What to build:** 缺陷的整体修改、局部字段修改（fixUserId/bugLevel/bugEnv/bugType/bugSource）、流转、缺陷确认在业务事务内先读旧值、计算字段级 diff，写入变更明细。缺陷确认同时记录确认状态与缺陷状态的前后值。

**Blocked by:** 01 后端变更明细基础设施

**Status:** ready-for-agent

- [ ] 整体修改生成变化字段的明细（含修复人、关联需求标题、版本名、模块名翻译）
- [ ] 局部修改生成单字段明细
- [ ] 流转生成状态 from→to 明细；确认生成确认状态与缺陷状态明细
- [ ] gofmt / go vet / go build 通过
