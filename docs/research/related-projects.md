# SQL 改写与优化相关项目调研

## 调研目标

本调研用于判断哪些开源项目和研究系统能为本项目提供设计参考。本项目定位仍保持不变：面向 BI / 报表场景的规则驱动 SQL 自动改写工具，不做任意 SQL 自动优化器。

## 重点参考项目

| 项目 | 类型 | 主要价值 | 对本项目的启发 |
|---|---|---|---|
| QueryBooster | SQL 改写中间件 | 中间件截获 SQL、人工规则、查询日志、重写路径、Web 管理。 | 参考其 rules / queries / rewriting paths 模型，但不复制 GPL 源码。 |
| Apache Calcite | 规则和代价优化框架 | SQL parser、validator、relational algebra、rule-based / cost-based optimizer。 | 学习规则分层、规则优先级、关系代数表达；不在 Go 项目中直接引入 Java 框架。 |
| SQLGlot | SQL parser / transpiler / optimizer | 多方言 SQL AST、格式化、转译和部分优化能力。 | 可作为离线规则验证和 SQL 样本分析参考；Go 主实现不直接依赖 Python。 |
| pg_query_go / libpg_query | PostgreSQL parser | 使用 PostgreSQL 源码解析 SQL，适合生成 PostgreSQL AST 和 fingerprint。 | 后续 Go 版 AST 识别优先评估。需要处理 cgo、版本和部署复杂度。 |
| Coral | SQL translation / rewrite engine | 使用中间表示做 SQL 翻译、分析和改写。 | 可参考 IR 思路，但本项目第一阶段不做跨引擎 SQL 翻译。 |
| pg-rewrite-proxy | PostgreSQL rewrite proxy | PostgreSQL 反向代理，通过 Lua 改写 Query / Parse 消息。 | 参考协议代理边界，尤其 extended query protocol 和参数绑定限制。 |
| ProxySQL / MaxScale | 数据库代理 | 支持基于规则的查询 rewrite、routing、filter。 | 参考生产代理的规则管理、灰度和回退模型；不直接作为 YMatrix/PostgreSQL 第一方案。 |
| HypoPG / Dexter / index_advisor | 物理设计 advisor | 用 hypothetical index 评估索引收益，输出 DDL 建议。 | 支撑本项目 `advisor` 模块，只建议不执行 DDL。 |
| ReadySet | SQL caching layer | 通过代理和增量维护加速重复读查询。 | 对“历史复用”和“预计算/缓存”有启发，但它是缓存系统，不是 SQL 改写工具。 |
| pg_ivm | 增量物化视图 | 增量维护物化视图，减少重复计算。 | 对报表预聚合表/物化表建议有参考价值，不直接作为自动改写依赖。 |
| WeTune | 自动发现 rewrite rules | 自动发现并验证 SQL rewrite rules。 | 适合长期研究，不适合当前客户交付阶段直接引入。 |
| pg_stat_statements | PostgreSQL workload 统计扩展 | 统计 SQL 执行次数、平均耗时、总耗时、计划耗时等。 | 可作为后续导入真实 workload 的标准来源，但它通常不保留完整前端上下文。 |
| pgBadger | PostgreSQL 日志分析器 | 从数据库日志生成慢 SQL 和性能报表。 | 可作为导入慢 SQL 样本的离线入口，适合客户现场只能给日志的场景。 |
| PoWA | PostgreSQL Workload Analyzer | 采集、聚合多类 PostgreSQL 性能指标，并提供图表和调优建议。 | 参考其 workload 历史化和 per-query 指标组织方式，不直接替代本项目。 |
| pg_hint_plan / pg_plan_advsr | PostgreSQL hint / plan advisor | 通过 SQL 注释或历史计划信息影响 PostgreSQL 执行计划。 | 只能作为高级可选建议，不能作为默认改写策略，否则会引入计划绑定风险。 |

## 可借鉴设计模式

### 1. 规则不是纯字符串替换

QueryBooster、Calcite、SQLGlot、Coral 都说明一个方向：可靠 SQL 改写需要结构化表示。字符串替换只能用于非常严格的已知模板，不能承担通用 SQL 理解。

本项目下一步建议：

- 继续保留严格 fingerprint。
- 增加 AST summary。
- 规则匹配同时检查关系、过滤字段、聚合、分组、排序、limit、distinct、窗口函数等结构。
- 模板渲染仍采用白名单占位符。

### 2. 重写路径需要被审计

QueryBooster 的 `rewriting_paths` 思路很适合本项目。即使第一阶段只有一步改写，也应该记录：

- 输入 SQL hash。
- 命中规则 ID。
- 规则版本。
- 输出 SQL hash。
- 改写来源：`history`、`rule`、`original`。
- 回退原因。

这样后续 Web viewer 可以解释为什么某条 SQL 被改写或没被改写。

### 3. 代理接入要后置

pg-rewrite-proxy、ProxySQL、MaxScale 说明代理模式可行，但复杂度高。尤其 PostgreSQL / YMatrix 场景要处理：

- SSL。
- startup packet。
- simple query protocol。
- extended query protocol。
- prepared statement。
- bind 参数。
- transaction state。
- cursor。
- error fallback。

本项目不应过早进入协议代理。更稳妥路线是先做 SQLite 证据链和 HTTP 服务，再评估代理。

### 4. advisor 只能给建议，不能自动执行

HypoPG、Dexter、index_advisor 的共同价值是“低成本评估物理设计收益”。这适合本项目生成字段类型、索引、分布键、排序键、预聚合表建议。

本项目必须继续坚持：

- `advisor` 只输出建议。
- 每条建议包含风险、验证方式和回滚方案。
- 不自动执行 DDL。

### 5. 自动规则生成只适合长期方向

WeTune、LLM-R2、R-Bot、QUITE 这类项目说明自动规则发现和 LLM 辅助改写是研究热点，但当前客户交付不适合直接使用。

本项目可以长期保留两个方向：

- 从原 SQL / 优化 SQL pair 中生成候选规则。
- 用 LLM 生成候选改写，再用 validate、EXPLAIN 和人工审核过滤。

但短期不能让自动生成结果直接进入 `candidate`。

### 6. workload 采集要先标准化

pg_stat_statements、pgBadger、PoWA 都说明：真正能持续优化的系统，不能只靠人工贴一条 SQL。它需要把 SQL 样本、执行次数、平均耗时、最大耗时、扫描行数、时间窗口、来源应用和参数形态统一记录下来。

本项目下一步建议把输入来源统一成内部 `query_logs` 模型：

- CLI 手工输入。
- BI 前端触发后从后台日志截取。
- pg_stat_statements 导出。
- 数据库慢 SQL 日志。
- 未来 HTTP / proxy 捕获。

这样不管 SQL 是从哪里来的，后续 fingerprint、history、rewrite、validate、advisor 都走同一套流程。

### 7. hint 不是主路线

pg_hint_plan 和 pg_plan_advsr 证明“在 SQL 中加入 hint”能影响 PostgreSQL 执行计划，但这类方案对数据库版本、插件安装、统计信息、数据分布变化都敏感。

本项目可以输出 hint 类建议，但不建议默认把 hint 注入到生产 SQL。更稳妥的策略仍然是：

- 优先改写 SQL 结构。
- 其次建议字段类型、索引、分布键、排序键、预聚合表。
- 最后才把 hint 作为需要 DBA 确认的高级建议。

## 对本项目的优先级建议

### P0：马上参考

- SQLite 元数据存储。
- `query_logs`。
- `rewriting_paths`。
- `validation_reports`。
- `rewrite_suggestions`。
- 统一 workload 导入模型。
- Web viewer 展示真实日志和重写路径。

这些直接服务当前“历史复用”和“客户可验收证据链”。

### P1：下一阶段评估

- `pg_query_go` 或其它 PostgreSQL parser，用于 AST summary 和更可靠 fingerprint。
- HypoPG / index advisor 类能力，用于 DDL 和物理设计建议。
- pg_stat_statements / 慢 SQL 日志导入器。
- HTTP 服务接口，复用当前 engine。

### P2：中长期观察

- PostgreSQL 协议代理。
- JDBC / Driver 侧截获。
- 自动规则生成。
- LLM 辅助 SQL 改写。
- 增量物化视图或缓存层。

## 不建议当前采用

- 直接复制 QueryBooster 代码：GPL-3.0，且技术栈是 Python / Flask / React。
- 直接引入 Apache Calcite：Java 框架重，和 Go 单二进制目标不一致。
- 直接做通用 SQLGlot 风格多方言转译：偏离当前客户场景。
- 直接上 ProxySQL / MaxScale：更偏 MySQL / MariaDB 生态，和 YMatrix/PostgreSQL 协议兼容目标不完全一致。
- 直接上 LLM 自动改写：结果一致性和性能稳定性不可控。
- 默认注入 pg_hint_plan hint：短期效果可能明显，但长期维护风险高。

## 推荐路线更新

下一阶段建议实施顺序：

1. 实现 SQLite store：`rules`、`history_records`、`query_logs`、`rewriting_paths`。
2. CLI 每次 `rewrite` 都落 query log。
3. history 命中也要落日志，但不重复生成改写路径。
4. Web viewer 从导出的 JSON 或只读 SQLite 查询结果读取真实数据。
5. 加入 AST summary，为未来安全规则匹配做准备。
6. 增加 pg_stat_statements / 慢 SQL 日志导入器。
7. 再评估 HTTP 服务和 PostgreSQL 协议代理。

## 参考链接

- QueryBooster：<https://github.com/ISG-ICS/QueryBooster>
- Apache Calcite：<https://calcite.apache.org/>
- SQLGlot：<https://github.com/tobymao/sqlglot>
- pg_query_go：<https://github.com/pganalyze/pg_query_go>
- libpg_query：<https://github.com/pganalyze/libpg_query>
- Coral：<https://github.com/linkedin/coral>
- pg-rewrite-proxy：<https://github.com/patientsknowbest/pg-rewrite-proxy>
- ProxySQL Query Rewrite：<https://proxysql.com/documentation/query-rewrite/>
- MaxScale Rewrite Filter：<https://mariadb.com/docs/maxscale/reference/maxscale-filters/maxscale-rewrite-filter>
- HypoPG：<https://github.com/HypoPG/hypopg>
- Dexter：<https://github.com/ankane/dexter>
- Supabase index_advisor：<https://github.com/supabase/index_advisor>
- ReadySet：<https://github.com/readysettech/readyset>
- pg_ivm：<https://github.com/sraoss/pg_ivm>
- WeTune：<https://github.com/WeTune/WeTune-code>
- pg_stat_statements：<https://www.postgresql.org/docs/current/pgstatstatements.html>
- pgBadger：<https://github.com/darold/pgbadger>
- PoWA：<https://powa.readthedocs.io/>
- pg_hint_plan：<https://github.com/ossc-db/pg_hint_plan>
- pg_plan_advsr：<https://github.com/ossc-db/pg_plan_advsr>
