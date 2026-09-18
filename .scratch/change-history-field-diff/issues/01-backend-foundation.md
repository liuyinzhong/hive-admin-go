# 01: 后端变更明细基础设施

**What to build:** 变更记录支持字段级明细的存储与返回：数据模型新增 change_items 列（含 up/down 迁移脚本，仅创建不执行）、按业务类型维护的字段目录注册表（字段→中文标签→字典类型→取值方式）、变更记录写入公共函数支持明细参数、查询接口响应返回 changeItems、Swagger 注释同步。

**Blocked by:** None (can start immediately)

**Status:** ready-for-agent

- [ ] dev_change_history 新增 change_items text 列，迁移脚本含回滚
- [ ] 明细 JSON 结构 {fieldKey, fieldLabel, dictType, oldValue, newValue}：字典值存原始值并携带 dictType，引用值（人名/版本名/模块名/需求标题/文件名）与 fieldLabel 由后端写入时翻译固化
- [ ] 字段目录注册表覆盖需求/任务/缺陷的可变更字段，含富文本字段"描述已更新"规则与附件增删文件名规则
- [ ] 公共写入函数（createChangeHistoryTx / updateDevRecordWithHistory）支持传入明细
- [ ] GET /dev/changeHistory 响应含 changeItems，空明细时为空数组
- [ ] Swagger 注释与响应 DTO 同步
- [ ] gofmt / go vet / go build 通过
