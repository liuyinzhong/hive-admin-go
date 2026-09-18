# 05: 工作流自动化接入变更明细

**What to build:** 工作流自动化"修改字段值"动作除现有富文本正文（来源流程、节点、动作名、旧值→新值描述）外，同步生成结构化字段级明细。插入记录动作保持现状。

**Blocked by:** 01 后端变更明细基础设施

**Status:** ready-for-agent

- [ ] 自动化修改字段值生成明细（fieldKey/fieldLabel/dictType/旧值/新值），正文保留来源描述
- [ ] gofmt / go vet / go build 通过
