# migrations 目录规则

本目录存放迁移 SQL。新增或修改本目录文件前必须先读本文件；数据库结构变更本身的兼容性、数据风险与回滚说明要求见上级 `AGENTS.md` 的「数据库结构变更」。

迁移 SQL 统一放在仓库根目录 `migrations/` 下，按环境和分支组织。脚本文件名统一沿用 `日期_序号_描述_up/down.sql` 格式，例如 `20260813_03_data_permission_owner_backfill_up.sql`。

`migrations/` 根目录存量 `*_up.sql`、`*_down.sql` 为已发布的历史迁移，不因本规则移动、删除或改名；本规则只约束今后新增迁移 SQL 的位置。

## dev：开发环境草稿

开发环境的迁移草稿、回滚草稿和数据修复草稿放入 `migrations/dev/<分支名>/`，按 `YYYYMMDD_NN_描述_up.sql` 与 `YYYYMMDD_NN_描述_down.sql` 成对命名；目录或文件不存在时先创建。

- 序号 `NN` 在同一天内从 `01` 递增；描述用小写英文单词和下划线概括变更主题。
- 每个变更主题独立一对文件，不同主题不得混入同一文件；同一天多个主题按序号新增文件。
- 未发布的草稿可直接修改完善自身文件，不回写其它文件。

示例：分支 `perf-2.43.0-09月优化与修复` 的一对文件为

```text
migrations/dev/perf-2.43.0-09月优化与修复/20260914_01_add_erp_order_table_up.sql
migrations/dev/perf-2.43.0-09月优化与修复/20260914_01_add_erp_order_table_down.sql
```

分支名通过 `git branch --show-current` 获取；分支名中如出现 `/` 等路径分隔字符，先替换为 `-`；无法确定分支名时（例如 detached HEAD），先说明情况并与用户确认命名。

## prod：面向生产的汇总迁移

默认只生成和更新 `migrations/dev/` 下的文件。`migrations/prod/` 下是该分支面向生产的**汇总**迁移，不再保留一个个独立脚本：成对生成 `<分支名>_up.sql` 和 `<分支名>_down.sql`，例如 `migrations/prod/perf-2.43.0-09月优化与修复_up.sql` 与 `perf-2.43.0-09月优化与修复_down.sql`。

- `_up.sql` 由 dev 下该分支全部 `_up.sql` 按文件名（日期、序号）升序合并而成。
- `_down.sql` 由对应 `_down.sql` 按倒序（后执行的变更先回滚）合并而成。
- 只在版本发布、用户明确说明生成到 prod 时才创建，平时不得提前生成或改动。
- 生成 prod 汇总文件时，必须在交付说明中标明用途、兼容性和回滚方式。
