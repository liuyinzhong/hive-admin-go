# 02: 需求写入点接入变更明细

**What to build:** 需求的整体修改、局部字段修改（storyType/storyLevel/source）、流转、批量流转在业务事务内先读旧值、计算字段级 diff，写入变更明细。富文本描述记"描述已更新"，附件记增删文件名，未变化字段不出现在明细中。

**Blocked by:** 01 后端变更明细基础设施

**Status:** ready-for-agent

- [ ] 整体修改生成变化字段的明细（含引用值翻译：版本名、模块名等）
- [ ] 局部修改生成单字段明细
- [ ] 流转与批量流转生成状态 from→to 明细，正文行为不变
- [ ] gofmt / go vet / go build 通过
