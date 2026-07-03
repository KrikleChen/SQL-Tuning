const runs = [
  {
    id: "SQL_001",
    table: "orders",
    source: "history",
    decision: "candidate",
    status: "passed",
    latencyBefore: 1840,
    latencyAfter: 420,
    reason: "history hit: raw SQL and optimizer version matched",
    original:
      "select * from report_orders where customer_key = 'CUSTOMER_001' order by created_at desc",
    rewritten:
      "select order_key, customer_key, status_code, total_amount, created_at from report_orders where customer_key = 'CUSTOMER_001' order by created_at desc limit 200",
    signals: ["历史记录验证通过", "指纹 sha256:8f21...a90b", "收益 77.2%"],
  },
  {
    id: "SQL_002",
    table: "order_items",
    source: "rule",
    decision: "candidate",
    status: "passed",
    latencyBefore: 2320,
    latencyAfter: 930,
    reason: "candidate produced by projection pruning rule",
    original:
      "select * from order_items where order_id in (select id from orders where status = 'paid')",
    rewritten:
      "select order_id, sku_id, quantity, price from order_items where order_id in (select id from orders where status = 'paid')",
    signals: ["规则命中: avoid_select_star", "字段裁剪", "收益 59.9%"],
  },
  {
    id: "SQL_003",
    table: "inventory_snapshot",
    source: "fallback",
    decision: "needs_review",
    status: "review",
    latencyBefore: 760,
    latencyAfter: 760,
    reason: "fallback: no reusable history and no safe rule matched",
    original:
      "select sku_id, warehouse_id, count(*) from inventory_snapshot group by sku_id, warehouse_id",
    rewritten:
      "select sku_id, warehouse_id, count(*) from inventory_snapshot group by sku_id, warehouse_id",
    signals: ["无历史命中", "未生成候选", "保持原 SQL"],
  },
  {
    id: "SQL_004",
    table: "account_activity",
    source: "history",
    decision: "reject",
    status: "failed",
    latencyBefore: 610,
    latencyAfter: 840,
    reason: "history hit rejected because stored version is not current",
    original:
      "select account_key, login_name_masked, last_login_at from account_activity where last_login_at is not null",
    rewritten:
      "select account_key, login_name_masked, last_login_at from account_activity where last_login_at is not null",
    signals: ["历史版本失效", "候选拒绝", "需要重新评估"],
  },
  {
    id: "SQL_005",
    table: "payments",
    source: "rule",
    decision: "candidate",
    status: "passed",
    latencyBefore: 1280,
    latencyAfter: 510,
    reason: "candidate produced by predicate tightening rule",
    original:
      "select id, order_id, amount from payments where created_at >= '2026-07-01'",
    rewritten:
      "select id, order_id, amount from payments where created_at >= '2026-07-01' and status in ('paid', 'settled')",
    signals: ["规则命中: tighten_payment_status", "扫描行数下降", "收益 60.2%"],
  },
  {
    id: "SQL_006",
    table: "click_events",
    source: "rule",
    decision: "candidate",
    status: "review",
    latencyBefore: 4980,
    latencyAfter: 2120,
    reason: "candidate produced but requires semantic validation",
    original:
      "select subject_key, count(*) from click_events where dt = '2026-07-02' group by subject_key",
    rewritten:
      "select subject_key, count(*) from click_events where dt = '2026-07-02' and subject_key is not null group by subject_key",
    signals: ["候选待验收", "空值语义需确认", "预计收益 57.4%"],
  },
  {
    id: "SQL_007",
    table: "refunds",
    source: "fallback",
    decision: "not_in_scope",
    status: "review",
    latencyBefore: 940,
    latencyAfter: 940,
    reason: "fallback: DML statement is outside rewrite scope",
    original: "update refunds set reviewed = 1 where created_at < '2026-01-01'",
    rewritten: "update refunds set reviewed = 1 where created_at < '2026-01-01'",
    signals: ["非查询语句", "跳过改写", "保持原 SQL"],
  },
  {
    id: "SQL_008",
    table: "shipments",
    source: "history",
    decision: "candidate",
    status: "passed",
    latencyBefore: 1560,
    latencyAfter: 600,
    reason: "history hit: normalized fingerprint matched",
    original:
      "select * from shipments where carrier_code = 'CARRIER_001' and shipped_at >= '2026-06-01'",
    rewritten:
      "select shipment_key, order_key, carrier_code, tracking_no_masked, shipped_at from shipments where carrier_code = 'CARRIER_001' and shipped_at >= '2026-06-01'",
    signals: ["规范化指纹命中", "字段裁剪已验收", "收益 61.5%"],
  },
];

const filters = [
  { key: "all", label: "全部" },
  { key: "history", label: "History" },
  { key: "candidate", label: "Candidate" },
  { key: "fallback", label: "Fallback" },
  { key: "review", label: "需复核" },
];

const labels = {
  history: "History",
  rule: "Rule",
  fallback: "Fallback",
  candidate: "Candidate",
  reject: "Reject",
  needs_review: "Needs Review",
  not_in_scope: "Not In Scope",
  passed: "Passed",
  failed: "Failed",
  review: "Review",
};

let activeFilter = "all";
let selectedId = runs[0].id;

const metricsEl = document.getElementById("metrics");
const filtersEl = document.getElementById("filters");
const runsEl = document.getElementById("runs");
const detailsEl = document.getElementById("details");
const searchEl = document.getElementById("search");

function escapeHTML(value) {
  return String(value)
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#039;");
}

function countBy(predicate) {
  return runs.filter(predicate).length;
}

function improvement(run) {
  if (run.latencyBefore <= 0) {
    return 0;
  }
  return Math.round(
    ((run.latencyBefore - run.latencyAfter) / run.latencyBefore) * 1000,
  ) / 10;
}

function renderMetrics() {
  const reviewCount = countBy((run) => run.status === "review");
  const metricItems = [
    ["样本数", runs.length, "本地 demo 数据"],
    ["History 命中", countBy((run) => run.source === "history"), "含版本校验"],
    ["Candidate", countBy((run) => run.decision === "candidate"), "可进入验收"],
    ["Fallback", countBy((run) => run.source === "fallback"), "保持原 SQL"],
    ["需复核", reviewCount, "人工确认队列"],
  ];

  metricsEl.innerHTML = metricItems
    .map(
      ([label, value, note]) => `
        <article class="metric">
          <span>${escapeHTML(label)}</span>
          <strong>${escapeHTML(value)}</strong>
          <small>${escapeHTML(note)}</small>
        </article>
      `,
    )
    .join("");
}

function renderFilters() {
  filtersEl.innerHTML = filters
    .map(
      (filter) => `
        <button
          class="tab"
          type="button"
          role="tab"
          aria-selected="${filter.key === activeFilter}"
          data-filter="${escapeHTML(filter.key)}"
        >${escapeHTML(filter.label)}</button>
      `,
    )
    .join("");
}

function matchesFilter(run) {
  if (activeFilter === "all") {
    return true;
  }
  if (activeFilter === "candidate") {
    return run.decision === "candidate";
  }
  if (activeFilter === "review") {
    return run.status === "review";
  }
  return run.source === activeFilter;
}

function matchesSearch(run) {
  const term = searchEl.value.trim().toLowerCase();
  if (!term) {
    return true;
  }
  return [run.id, run.table, run.source, run.decision, run.status, run.reason]
    .join(" ")
    .toLowerCase()
    .includes(term);
}

function visibleRuns() {
  return runs.filter((run) => matchesFilter(run) && matchesSearch(run));
}

function renderRuns() {
  const rows = visibleRuns();
  if (!rows.length) {
    runsEl.innerHTML = `
      <tr class="empty-row">
        <td colspan="6">没有匹配的样本</td>
      </tr>
    `;
    return;
  }

  if (!rows.some((run) => run.id === selectedId)) {
    selectedId = rows[0].id;
  }

  runsEl.innerHTML = rows
    .map((run) => {
      const gain = improvement(run);
      const deltaClass = gain > 0 ? "delta" : "delta warn";
      return `
        <tr
          tabindex="0"
          class="${run.id === selectedId ? "is-selected" : ""}"
          data-id="${escapeHTML(run.id)}"
        >
          <td>
            <span class="sql-id">${escapeHTML(run.id)}</span>
            <span class="table-name">${escapeHTML(run.table)}</span>
          </td>
          <td><span class="tag ${escapeHTML(run.source)}">${escapeHTML(labels[run.source])}</span></td>
          <td><span class="tag ${escapeHTML(run.decision)}">${escapeHTML(labels[run.decision])}</span></td>
          <td><span class="${deltaClass}">${gain > 0 ? "-" : ""}${escapeHTML(Math.abs(gain))}%</span></td>
          <td><span class="tag ${escapeHTML(run.status)}">${escapeHTML(labels[run.status])}</span></td>
          <td class="note">${escapeHTML(run.reason)}</td>
        </tr>
      `;
    })
    .join("");
}

function renderDetails() {
  const rows = visibleRuns();
  if (!rows.length) {
    detailsEl.innerHTML = `
      <div class="detail-header">
        <div class="detail-title">
          <div>
            <strong>没有匹配的样本详情</strong>
            <p class="table-name">调整筛选或搜索条件</p>
          </div>
        </div>
      </div>
    `;
    return;
  }

  const run = rows.find((item) => item.id === selectedId) || rows[0];
  if (!run) {
    detailsEl.innerHTML = "";
    return;
  }

  selectedId = run.id;
  const gain = improvement(run);
  detailsEl.innerHTML = `
    <div class="detail-header">
      <div class="detail-title">
        <div>
          <strong>${escapeHTML(run.id)}</strong>
          <p class="table-name">${escapeHTML(run.table)}</p>
        </div>
        <span class="tag ${escapeHTML(run.source)}">${escapeHTML(labels[run.source])}</span>
      </div>
      <div class="detail-grid">
        <div class="detail-stat">
          <span>原耗时</span>
          <strong>${escapeHTML(run.latencyBefore)} ms</strong>
        </div>
        <div class="detail-stat">
          <span>当前耗时</span>
          <strong>${escapeHTML(run.latencyAfter)} ms</strong>
        </div>
        <div class="detail-stat">
          <span>收益</span>
          <strong>${escapeHTML(gain)}%</strong>
        </div>
        <div class="detail-stat">
          <span>状态</span>
          <strong>${escapeHTML(labels[run.status])}</strong>
        </div>
      </div>
    </div>
    <div class="detail-body">
      <section class="sql-block">
        <h3>原 SQL</h3>
        <pre>${escapeHTML(run.original)}</pre>
      </section>
      <section class="sql-block">
        <h3>候选 / 回退 SQL</h3>
        <pre>${escapeHTML(run.rewritten)}</pre>
      </section>
      <section class="sql-block">
        <h3>判定信号</h3>
        <ul class="reason-list">
          ${run.signals.map((signal) => `<li>${escapeHTML(signal)}</li>`).join("")}
        </ul>
      </section>
    </div>
  `;
}

function render() {
  renderMetrics();
  renderFilters();
  renderRuns();
  renderDetails();
}

filtersEl.addEventListener("click", (event) => {
  const button = event.target.closest("[data-filter]");
  if (!button) {
    return;
  }
  activeFilter = button.dataset.filter;
  render();
});

runsEl.addEventListener("click", (event) => {
  const row = event.target.closest("[data-id]");
  if (!row) {
    return;
  }
  selectedId = row.dataset.id;
  renderRuns();
  renderDetails();
});

runsEl.addEventListener("keydown", (event) => {
  if (event.key !== "Enter" && event.key !== " ") {
    return;
  }
  const row = event.target.closest("[data-id]");
  if (!row) {
    return;
  }
  event.preventDefault();
  selectedId = row.dataset.id;
  renderRuns();
  renderDetails();
});

searchEl.addEventListener("input", () => {
  renderRuns();
  renderDetails();
});

render();
