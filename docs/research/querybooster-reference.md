# QueryBooster 参考调研

## 参考来源

- GitHub：<https://github.com/ISG-ICS/QueryBooster>
- VLDB 2023 论文：<https://www.vldb.org/pvldb/vol16/p2911-bai.pdf>
- VLDB 2023 Demo 论文：<https://www.vldb.org/pvldb/vol16/p4038-bai.pdf>

## 结论

QueryBooster 的核心思路值得参考：它把 SQL 改写放在应用和数据库之间，由中间件截获 BI 工具发出的 SQL，再基于人工规则做语义等价改写。这个定位和本项目一致：规则驱动、人参与、可解释、可审计，而不是任意 SQL 自动优化器。

本项目只参考架构思想，不复制源码。QueryBooster 使用 GPL-3.0，本项目后续实现必须保持自研。

## 可借鉴点

| QueryBooster 思路 | 本项目对应设计 |
|---|---|
| 中间件截获 SQL | 后续 HTTP 服务、协议代理或 JDBC 方案评估。 |
| 人工规则改写 | 保持已验证模板和严格规则库。 |
| 规则包含 pattern / rewrite / constraints / actions | 后续扩展规则元数据，但第一阶段不做通用规则语言。 |
| 保存 queries 和 rewriting_paths | 第二阶段实现 SQLite query logs 和 rewriting paths。 |
| 规则启用绑定应用 | 后续支持按项目、报表、数据库和环境启停规则。 |
| 建议规则和候选规则 | 本项目用 `rewrite_suggestions` 保存未进入自动候选库的建议。 |
| Web 查看查询日志和改写路径 | 本项目静态 viewer 后续读取 SQLite 导出结果。 |

## 不照搬点

- 不复制 Python / Flask / React 代码。
- 不直接采用 GPL 代码片段。
- 不在当前阶段实现 JDBC Driver 定制。
- 不在当前阶段实现自动规则生成。
- 不在当前阶段实现完整规则管理平台。
- 不放宽本项目的 `candidate + validate passed + schema/template/rule version` 门禁。

## 对后续路线的影响

原路线是 CLI 后直接考虑 HTTP 服务和协议代理。参考 QueryBooster 后，路线调整为：

1. CLI + engine + history。
2. SQLite 元数据、查询日志和重写路径。
3. 本地 Web viewer 读取真实产物。
4. HTTP 服务。
5. 协议代理或 JDBC 方案评估。
6. 规则管理平台。

这个顺序更适合当前项目：先把可审计、可复用、可验收的证据链做扎实，再接线上流量。
