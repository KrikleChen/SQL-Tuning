# SQLOpt Framework Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the first runnable framework for a Go-based SQL rewrite tool with history reuse, high-contrast CLI output, and a lightweight local web viewer.

**Architecture:** The CLI is a thin entry point over reusable internal packages. The backend first implements deterministic identity, history lookup, output formatting, and a minimal rewrite engine. The frontend is a static local viewer for audit/report artifacts and does not control rewrite behavior.

**Tech Stack:** Go 1.22, Go standard library, static HTML/CSS/JavaScript.

---

## File Structure

- `go.mod`: Go module declaration.
- `cmd/sqlopt/main.go`: CLI entry point.
- `internal/output/output.go`: High-contrast status formatter, color policy, stdout/stderr separation helpers.
- `internal/output/output_test.go`: Tests that black and dark gray ANSI codes are never emitted for status messages.
- `internal/normalize/normalize.go`: SQL normalization for whitespace and trimming.
- `internal/normalize/normalize_test.go`: Normalization behavior tests.
- `internal/fingerprint/fingerprint.go`: SHA256 fingerprint generation.
- `internal/fingerprint/fingerprint_test.go`: Stable fingerprint tests.
- `internal/history/history.go`: History record, lookup keys, in-memory store, reusable decision checks.
- `internal/history/history_test.go`: Exact hit, fingerprint hit, version invalidation, and non-candidate rejection tests.
- `internal/engine/engine.go`: Rewrite request/result and orchestration over history, identity, and static rules.
- `internal/engine/engine_test.go`: History-first behavior, fallback behavior, and deterministic rewrite behavior tests.
- `web/index.html`: Static local viewer shell.
- `web/styles.css`: High-contrast UI styles.
- `web/app.js`: Demo data rendering for rules, history hits, and decisions.
- `README.md`: Basic commands and current scope.

---

### Task 1: Backend Foundation and CLI Output

**Files:**
- Create: `go.mod`
- Create: `cmd/sqlopt/main.go`
- Create: `internal/output/output.go`
- Create: `internal/output/output_test.go`

- [ ] **Step 1: Write failing output tests**

Create `internal/output/output_test.go`:

```go
package output

import (
	"strings"
	"testing"
)

func TestStatusMessageNeverUsesBlackOrDarkGray(t *testing.T) {
	msg := FormatStatus(StatusInfo, "history hit", ColorAlways)
	if strings.Contains(msg, "\x1b[30m") || strings.Contains(msg, "\x1b[90m") {
		t.Fatalf("status message uses unreadable dark ANSI color: %q", msg)
	}
}

func TestStatusMessageIncludesTextLabel(t *testing.T) {
	msg := FormatStatus(StatusHistory, "reused optimized SQL", ColorNever)
	if !strings.Contains(msg, "[HISTORY]") {
		t.Fatalf("expected history label, got %q", msg)
	}
}

func TestColorNeverDisablesANSI(t *testing.T) {
	msg := FormatStatus(StatusError, "failed", ColorNever)
	if strings.Contains(msg, "\x1b[") {
		t.Fatalf("expected no ANSI color when color disabled, got %q", msg)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/output`

Expected: FAIL because package and functions do not exist.

- [ ] **Step 3: Implement output formatter and CLI skeleton**

Create `go.mod`:

```go
module github.com/KrikleChen/SQL-Tuning

go 1.22
```

Create `internal/output/output.go`:

```go
package output

import "fmt"

type ColorMode string

const (
	ColorAuto   ColorMode = "auto"
	ColorAlways ColorMode = "always"
	ColorNever  ColorMode = "never"
)

type Status string

const (
	StatusOK       Status = "ok"
	StatusInfo     Status = "info"
	StatusWarn     Status = "warn"
	StatusError    Status = "error"
	StatusFallback Status = "fallback"
	StatusHistory  Status = "history"
)

func FormatStatus(status Status, message string, color ColorMode) string {
	label, ansi := statusStyle(status)
	text := fmt.Sprintf("%s %s", label, message)
	if color == ColorNever {
		return text
	}
	return ansi + text + "\x1b[0m"
}

func statusStyle(status Status) (string, string) {
	switch status {
	case StatusOK:
		return "[OK]", "\x1b[32m"
	case StatusWarn:
		return "[WARN]", "\x1b[33m"
	case StatusError:
		return "[ERROR]", "\x1b[31m"
	case StatusFallback:
		return "[FALLBACK]", "\x1b[33m"
	case StatusHistory:
		return "[HISTORY]", "\x1b[36m"
	default:
		return "[INFO]", "\x1b[36m"
	}
}
```

Create `cmd/sqlopt/main.go`:

```go
package main

import (
	"fmt"
	"os"

	"github.com/KrikleChen/SQL-Tuning/internal/output"
)

func main() {
	fmt.Fprintln(os.Stderr, output.FormatStatus(output.StatusInfo, "sqlopt framework initialized", output.ColorAlways))
}
```

- [ ] **Step 4: Run tests**

Run: `go test ./...`

Expected: PASS.

---

### Task 2: SQL Identity Packages

**Files:**
- Create: `internal/normalize/normalize.go`
- Create: `internal/normalize/normalize_test.go`
- Create: `internal/fingerprint/fingerprint.go`
- Create: `internal/fingerprint/fingerprint_test.go`

- [ ] **Step 1: Write failing normalization and fingerprint tests**

Create `internal/normalize/normalize_test.go`:

```go
package normalize

import "testing"

func TestNormalizeCollapsesWhitespace(t *testing.T) {
	got := SQL(" select  *\nfrom   t  ")
	want := "select * from t"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
```

Create `internal/fingerprint/fingerprint_test.go`:

```go
package fingerprint

import "testing"

func TestSHA256IsStable(t *testing.T) {
	a := SHA256("select * from t")
	b := SHA256("select * from t")
	if a != b {
		t.Fatalf("fingerprint not stable: %q vs %q", a, b)
	}
	if len(a) != len("sha256:")+64 {
		t.Fatalf("unexpected fingerprint length: %q", a)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/normalize ./internal/fingerprint`

Expected: FAIL because functions do not exist.

- [ ] **Step 3: Implement minimal identity packages**

Create `internal/normalize/normalize.go`:

```go
package normalize

import "strings"

func SQL(input string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(input)), " ")
}
```

Create `internal/fingerprint/fingerprint.go`:

```go
package fingerprint

import (
	"crypto/sha256"
	"encoding/hex"
)

func SHA256(input string) string {
	sum := sha256.Sum256([]byte(input))
	return "sha256:" + hex.EncodeToString(sum[:])
}
```

- [ ] **Step 4: Run tests**

Run: `go test ./...`

Expected: PASS.

---

### Task 3: History Store and Reuse Gate

**Files:**
- Create: `internal/history/history.go`
- Create: `internal/history/history_test.go`

- [ ] **Step 1: Write failing history tests**

Create `internal/history/history_test.go`:

```go
package history

import "testing"

func TestLookupExactReturnsOnlyReusableCandidate(t *testing.T) {
	store := NewMemoryStore()
	record := Record{
		ID:                    "hist_001",
		RawSQLHash:            "raw",
		NormalizedFingerprint: "fp",
		ParamsHash:            "params",
		RuleID:                "rule",
		RuleVersion:           1,
		TemplateChecksum:      "tmpl",
		SchemaSignature:       "schema",
		RewrittenSQL:          "select 1",
		Decision:              DecisionCandidate,
		ValidateStatus:        ValidatePassed,
		RuleEnabled:           true,
	}
	store.Save(record)

	got, ok := store.LookupExact(Query{
		RawSQLHash:       "raw",
		RuleVersion:      1,
		TemplateChecksum: "tmpl",
		SchemaSignature:  "schema",
	})
	if !ok || got.RewrittenSQL != "select 1" {
		t.Fatalf("expected reusable history hit, got %#v ok=%v", got, ok)
	}
}

func TestLookupRejectsVersionMismatch(t *testing.T) {
	store := NewMemoryStore()
	store.Save(Record{
		ID:                    "hist_001",
		RawSQLHash:            "raw",
		NormalizedFingerprint: "fp",
		ParamsHash:            "params",
		RuleID:                "rule",
		RuleVersion:           1,
		TemplateChecksum:      "tmpl",
		SchemaSignature:       "schema",
		RewrittenSQL:          "select 1",
		Decision:              DecisionCandidate,
		ValidateStatus:        ValidatePassed,
		RuleEnabled:           true,
	})

	_, ok := store.LookupExact(Query{
		RawSQLHash:       "raw",
		RuleVersion:      2,
		TemplateChecksum: "tmpl",
		SchemaSignature:  "schema",
	})
	if ok {
		t.Fatal("expected history miss when rule version changes")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/history`

Expected: FAIL because package does not exist.

- [ ] **Step 3: Implement memory history store**

Create `internal/history/history.go`:

```go
package history

type Decision string

const (
	DecisionCandidate Decision = "candidate"
)

type ValidateStatus string

const (
	ValidatePassed ValidateStatus = "passed"
)

type Record struct {
	ID                    string
	RawSQLHash            string
	NormalizedFingerprint string
	ParamsHash            string
	RuleID                string
	RuleVersion           int
	TemplateChecksum      string
	SchemaSignature       string
	RewrittenSQL          string
	Decision              Decision
	ValidateStatus        ValidateStatus
	RuleEnabled           bool
}

type Query struct {
	RawSQLHash            string
	NormalizedFingerprint string
	ParamsHash            string
	RuleVersion           int
	TemplateChecksum      string
	SchemaSignature       string
}

type MemoryStore struct {
	records []Record
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{}
}

func (s *MemoryStore) Save(record Record) {
	s.records = append(s.records, record)
}

func (s *MemoryStore) LookupExact(query Query) (Record, bool) {
	for _, record := range s.records {
		if record.RawSQLHash == query.RawSQLHash && reusable(record, query) {
			return record, true
		}
	}
	return Record{}, false
}

func (s *MemoryStore) LookupFingerprint(query Query) (Record, bool) {
	for _, record := range s.records {
		if record.NormalizedFingerprint == query.NormalizedFingerprint &&
			record.ParamsHash == query.ParamsHash &&
			reusable(record, query) {
			return record, true
		}
	}
	return Record{}, false
}

func reusable(record Record, query Query) bool {
	return record.Decision == DecisionCandidate &&
		record.ValidateStatus == ValidatePassed &&
		record.RuleEnabled &&
		record.RuleVersion == query.RuleVersion &&
		record.TemplateChecksum == query.TemplateChecksum &&
		record.SchemaSignature == query.SchemaSignature
}
```

- [ ] **Step 4: Run tests**

Run: `go test ./...`

Expected: PASS.

---

### Task 4: Minimal Rewrite Engine

**Files:**
- Create: `internal/engine/engine.go`
- Create: `internal/engine/engine_test.go`

- [ ] **Step 1: Write failing engine tests**

Create `internal/engine/engine_test.go`:

```go
package engine

import (
	"testing"

	"github.com/KrikleChen/SQL-Tuning/internal/history"
)

func TestRewriteUsesHistoryBeforeRules(t *testing.T) {
	store := history.NewMemoryStore()
	store.Save(history.Record{
		ID:                    "hist_001",
		RawSQLHash:            HashRawSQL("select * from slow"),
		NormalizedFingerprint: FingerprintSQL("select * from slow"),
		ParamsHash:            "params:none",
		RuleID:                "rule",
		RuleVersion:           1,
		TemplateChecksum:      "tmpl",
		SchemaSignature:       "schema",
		RewrittenSQL:          "select * from fast",
		Decision:              history.DecisionCandidate,
		ValidateStatus:        history.ValidatePassed,
		RuleEnabled:           true,
	})

	e := New(store, Rule{
		ID:               "rule",
		Version:          1,
		TemplateChecksum: "tmpl",
		SchemaSignature:  "schema",
		RawSQL:           "select * from slow",
		RewrittenSQL:     "select * from rule",
	})

	got := e.Rewrite(Request{SQL: "select * from slow"})
	if got.Source != SourceHistory || got.SQL != "select * from fast" {
		t.Fatalf("expected history rewrite, got %#v", got)
	}
}

func TestRewriteFallsBackToOriginalWhenNoRuleMatches(t *testing.T) {
	e := New(history.NewMemoryStore())
	got := e.Rewrite(Request{SQL: "select * from unknown"})
	if got.Source != SourceOriginal || got.SQL != "select * from unknown" {
		t.Fatalf("expected original fallback, got %#v", got)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/engine`

Expected: FAIL because engine package does not exist.

- [ ] **Step 3: Implement minimal engine**

Create `internal/engine/engine.go`:

```go
package engine

import (
	"github.com/KrikleChen/SQL-Tuning/internal/fingerprint"
	"github.com/KrikleChen/SQL-Tuning/internal/history"
	"github.com/KrikleChen/SQL-Tuning/internal/normalize"
)

type Source string

const (
	SourceHistory  Source = "history"
	SourceRule     Source = "rule"
	SourceOriginal Source = "original"
)

type Request struct {
	SQL string
}

type Result struct {
	SQL    string
	Source Source
	Reason string
}

type Rule struct {
	ID               string
	Version          int
	TemplateChecksum string
	SchemaSignature  string
	RawSQL           string
	RewrittenSQL     string
}

type Engine struct {
	history *history.MemoryStore
	rules   []Rule
}

func New(store *history.MemoryStore, rules ...Rule) *Engine {
	return &Engine{history: store, rules: rules}
}

func (e *Engine) Rewrite(request Request) Result {
	for _, rule := range e.rules {
		query := history.Query{
			RawSQLHash:            HashRawSQL(request.SQL),
			NormalizedFingerprint: FingerprintSQL(request.SQL),
			ParamsHash:            "params:none",
			RuleVersion:           rule.Version,
			TemplateChecksum:      rule.TemplateChecksum,
			SchemaSignature:       rule.SchemaSignature,
		}
		if record, ok := e.history.LookupExact(query); ok {
			return Result{SQL: record.RewrittenSQL, Source: SourceHistory, Reason: "history_exact_hit"}
		}
		if FingerprintSQL(rule.RawSQL) == query.NormalizedFingerprint {
			return Result{SQL: rule.RewrittenSQL, Source: SourceRule, Reason: "rule_match"}
		}
	}
	return Result{SQL: request.SQL, Source: SourceOriginal, Reason: "no_match"}
}

func HashRawSQL(sql string) string {
	return fingerprint.SHA256(sql)
}

func FingerprintSQL(sql string) string {
	return fingerprint.SHA256(normalize.SQL(sql))
}
```

- [ ] **Step 4: Run tests**

Run: `go test ./...`

Expected: PASS.

---

### Task 5: Static Web Viewer

**Files:**
- Create: `web/index.html`
- Create: `web/styles.css`
- Create: `web/app.js`

- [ ] **Step 1: Create local viewer files**

Create `web/index.html`:

```html
<!doctype html>
<html lang="zh-CN">
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>SQL 优化工具控制台</title>
    <link rel="stylesheet" href="./styles.css">
  </head>
  <body>
    <main class="shell">
      <header class="topbar">
        <div>
          <h1>SQL 优化工具控制台</h1>
          <p>查看规则命中、历史复用、回退和验收状态。</p>
        </div>
      </header>
      <section class="metrics" id="metrics"></section>
      <section class="table-wrap">
        <h2>样本执行结果</h2>
        <table>
          <thead>
            <tr>
              <th>SQL ID</th>
              <th>来源</th>
              <th>决策</th>
              <th>状态</th>
              <th>说明</th>
            </tr>
          </thead>
          <tbody id="runs"></tbody>
        </table>
      </section>
    </main>
    <script src="./app.js"></script>
  </body>
</html>
```

Create `web/styles.css`:

```css
:root {
  color-scheme: light;
  --bg: #f6f8fb;
  --panel: #ffffff;
  --text: #17202a;
  --muted: #586575;
  --line: #d9e0ea;
  --ok: #147d43;
  --warn: #9a6700;
  --err: #b42318;
  --info: #0969da;
}

* {
  box-sizing: border-box;
}

body {
  margin: 0;
  background: var(--bg);
  color: var(--text);
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
}

.shell {
  max-width: 1120px;
  margin: 0 auto;
  padding: 28px;
}

.topbar {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 24px;
}

h1,
h2,
p {
  margin: 0;
}

h1 {
  font-size: 28px;
  line-height: 1.2;
}

p {
  margin-top: 8px;
  color: var(--muted);
}

.metrics {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
  margin-bottom: 24px;
}

.metric,
.table-wrap {
  background: var(--panel);
  border: 1px solid var(--line);
  border-radius: 8px;
}

.metric {
  padding: 16px;
}

.metric strong {
  display: block;
  font-size: 26px;
}

.table-wrap {
  overflow-x: auto;
}

.table-wrap h2 {
  padding: 16px;
  font-size: 18px;
}

table {
  width: 100%;
  border-collapse: collapse;
}

th,
td {
  padding: 12px 16px;
  border-top: 1px solid var(--line);
  text-align: left;
  white-space: nowrap;
}

th {
  color: var(--muted);
  font-size: 13px;
}

.tag {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  padding: 4px 8px;
  font-size: 12px;
  font-weight: 700;
}

.tag.history {
  background: #ddf4ff;
  color: var(--info);
}

.tag.candidate {
  background: #dafbe1;
  color: var(--ok);
}

.tag.fallback,
.tag.needs_review {
  background: #fff8c5;
  color: var(--warn);
}

@media (max-width: 760px) {
  .metrics {
    grid-template-columns: 1fr;
  }

  .shell {
    padding: 18px;
  }
}
```

Create `web/app.js`:

```javascript
const runs = [
  { id: "SQL_001", source: "history", decision: "candidate", status: "passed", note: "复用已验证高性能 SQL" },
  { id: "SQL_002", source: "rule", decision: "candidate", status: "passed", note: "规则命中并生成改写 SQL" },
  { id: "SQL_003", source: "original", decision: "not_in_scope", status: "not_in_scope", note: "未命中规则，保持原 SQL" },
];

const metrics = [
  ["样本数", runs.length],
  ["历史复用", runs.filter((run) => run.source === "history").length],
  ["候选 SQL", runs.filter((run) => run.decision === "candidate").length],
  ["回退", runs.filter((run) => run.source === "original").length],
];

document.getElementById("metrics").innerHTML = metrics
  .map(([label, value]) => `<div class="metric"><strong>${value}</strong><span>${label}</span></div>`)
  .join("");

document.getElementById("runs").innerHTML = runs
  .map((run) => `
    <tr>
      <td>${run.id}</td>
      <td><span class="tag ${run.source}">${run.source}</span></td>
      <td><span class="tag ${run.decision}">${run.decision}</span></td>
      <td>${run.status}</td>
      <td>${run.note}</td>
    </tr>
  `)
  .join("");
```

- [ ] **Step 2: Verify static files exist**

Run: `ls web`

Expected: `app.js`, `index.html`, `styles.css`.

---

### Task 6: README and Full Verification

**Files:**
- Create or modify: `README.md`

- [ ] **Step 1: Add current usage docs**

Create `README.md`:

```markdown
# SQL 优化工具

这是一个规则驱动的 BI 慢 SQL 自动改写工具框架。第一阶段提供 Go CLI、历史改写结果复用、高对比控制台输出和本地 Web 查看界面。

## 当前能力

- 生成 SQL 规范化结果和指纹。
- 优先复用已经验证通过的历史改写 SQL。
- 未命中时保持原 SQL。
- 控制台状态消息不使用黑色或深灰色，适配黑色背景终端。
- `web/` 提供本地静态查看界面。

## 开发验证

运行全部测试：

```bash
go test ./...
```

运行 CLI：

```bash
go run ./cmd/sqlopt
```

打开本地查看界面：

```bash
open web/index.html
```
```

- [ ] **Step 2: Run full verification**

Run: `go test ./...`

Expected: PASS.

Run: `go run ./cmd/sqlopt`

Expected: stderr contains `[INFO] sqlopt framework initialized` and does not contain ANSI black `\x1b[30m` or dark gray `\x1b[90m`.

Run: `git status --short`

Expected: only intended project files changed.
