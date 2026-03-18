package report

func Template() string {
	return `
<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<title>Sentra Report</title>

<style>
body {
  background: #0f1115;
  color: #e6e6e6;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  margin: 0;
}

.container {
  max-width: 1200px;
  margin: auto;
  padding: 20px;
}

/* HEADER */
.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid #2a2f3a;
  padding-bottom: 10px;
  margin-bottom: 20px;
}

.logo {
  font-weight: bold;
  font-size: 20px;
}

.logo span {
  color: #00ffd0;
}

.version {
  color: #888;
}

/* META */
.meta-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 10px;
  margin-bottom: 20px;
}

.card {
  background: #161a22;
  padding: 12px;
  border-radius: 6px;
}

.card-label {
  font-size: 11px;
  color: #888;
}

.card-value {
  font-size: 14px;
}

/* TOOLBAR */
.toolbar {
  display: flex;
  justify-content: space-between;
  margin: 15px 0;
}

.filter-btn {
  margin-right: 8px;
  cursor: pointer;
  font-size: 12px;
  color: #aaa;
}

.filter-btn.active {
  color: #fff;
  font-weight: bold;
}

.actions input {
  background: #111;
  border: 1px solid #333;
  color: #eee;
  padding: 4px 8px;
}

/* FINDINGS */
.finding {
  background: #161a22;
  border-radius: 6px;
  margin-bottom: 12px;
  overflow: hidden;
  border-left: 4px solid #555;
}

.finding.alert-card { border-color: #ff4d4f; }
.finding.warning-card { border-color: #faad14; }
.finding.notice-card { border-color: #1890ff; }

.finding-header {
  padding: 10px;
  cursor: pointer;
  display: flex;
  gap: 10px;
  align-items: center;
}

.finding-body {
  display: none;
  padding: 10px;
  border-top: 1px solid #2a2f3a;
}

.finding.expanded .finding-body {
  display: block;
}

.score-num {
  font-weight: bold;
}

.badge {
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 11px;
}

.badge.alert { background: #ff4d4f; }
.badge.warning { background: #faad14; }
.badge.notice { background: #1890ff; }

.file-type-badge {
  color: #00ffd0;
}

.rule-name {
  flex: 1;
}

/* KV GRID */
.kv-grid {
  display: grid;
  grid-template-columns: 120px 1fr;
  gap: 6px;
  font-size: 12px;
}

.kv-key {
  color: #888;
}

.hash {
  color: #fadb14;
}

.path {
  color: #69c0ff;
}

/* TAGS */
.atk-tag {
  display: inline-block;
  background: #262b36;
  padding: 2px 6px;
  margin: 2px;
  border-radius: 4px;
  font-size: 11px;
}

/* STRINGS */
.mstr {
  display: block;
  font-size: 11px;
  color: #aaa;
}

/* COPY */
.copy {
  cursor: pointer;
}
.copy:hover {
  color: #00ffd0;
}

/* REASON BOX */
.reason-box {
  margin-top: 10px;
  padding: 10px;
  background: #11151c;
  border-radius: 6px;
}

.reason-header {
  display: flex;
  gap: 10px;
  margin-bottom: 5px;
}

.subscore-tag {
  color: #faad14;
}

.sigtype-tag {
  color: #00ffd0;
}

.section-title {
  font-size: 13px;
  margin-top: 10px;
  margin-bottom: 5px;
  color: #aaa;
}
</style>
</head>

<body>
<div class="container">

<!-- HEADER -->
<div class="header">
  <div class="logo">SENTRA<span> //</span></div>
  <div class="version">v{{.Version}}</div>
  <div>{{.ScanID}} · {{.GeneratedAt}}</div>
</div>

<!-- META -->
<div class="meta-grid">
  <div class="card"><div class="card-label">Hostname</div><div class="card-value">{{.Hostname}}</div></div>
  <div class="card"><div class="card-label">IP</div><div class="card-value">{{.IPAddr}}</div></div>
  <div class="card"><div class="card-label">OS</div><div class="card-value">{{.OS}}</div></div>
  <div class="card"><div class="card-label">User</div><div class="card-value">{{.User}}</div></div>
</div>

<!-- SUMMARY -->
<div class="card">
  <div class="section-title">Scan Summary</div>
  <div class="kv-grid">
    <span class="kv-key">Target</span><span>{{.Target}}</span>
    <span class="kv-key">Start</span><span>{{.ScanStart}}</span>
    <span class="kv-key">End</span><span>{{.ScanEnd}}</span>
    <span class="kv-key">Duration</span><span>{{.Duration}}</span>
    <span class="kv-key">Scanned</span><span>{{.Scanned}}</span>
    <span class="kv-key">Matches</span><span>{{.MatchCount}}</span>
    <span class="kv-key">Errors</span><span>{{.ErrorCount}}</span>
  </div>
</div>

<!-- TOOLBAR -->
<div class="toolbar">
  <div class="filters">
    <span class="filter-btn active" data-filter="all">All ({{len .Findings}})</span>
    <span class="filter-btn alert" data-filter="alert">Alert ({{.AlertCount}})</span>
    <span class="filter-btn warning" data-filter="warning">Warning ({{.WarningCount}})</span>
    <span class="filter-btn notice" data-filter="notice">Notice ({{.NoticeCount}})</span>
  </div>

  <div class="actions">
    <input id="searchBox" placeholder="Search..." />
    <button onclick="expandAll()">Expand</button>
    <button onclick="collapseAll()">Collapse</button>
  </div>
</div>

<!-- FINDINGS -->
{{range .Findings}}
<div class="finding {{.Severity}}-card"
     data-severity="{{.Severity}}"
     data-text="{{.RuleName}} {{.Target}} {{.Description}}">

  <div class="finding-header">
    <span class="score-num" data-score="{{.Score}}">{{.Score}}</span>
    <span class="badge {{.Severity}}">{{.Severity}}</span>
    <span class="file-type-badge">{{.FileType}}</span>
    <span class="rule-name">{{.RuleName}}</span>
  </div>

  <div class="finding-body">

    <div class="section-title">File Info</div>
    <div class="kv-grid">
      <span class="kv-key">Target</span>
      <span class="kv-val path copy">{{.Target}}</span>

      {{if .MD5}}
      <span class="kv-key">MD5</span>
      <span class="kv-val hash copy">{{.MD5}}</span>
      {{end}}

      {{if .SHA256}}
      <span class="kv-key">SHA256</span>
      <span class="kv-val hash copy">{{.SHA256}}</span>
      {{end}}
    </div>

    <div class="reason-box">
      <div class="reason-header">
        <span>{{.RuleName}}</span>
        <span class="subscore-tag">Score: {{.Score}}</span>
        <span class="sigtype-tag">{{.RuleType}}</span>
      </div>

      <div>{{.Description}}</div>

      <div>
        Author: {{.Author}} · Date: {{.Date}} · Class: {{.Class}}
      </div>

      <div>
        {{range .AttackTags}}
        <span class="atk-tag">{{.}}</span>
        {{end}}
      </div>

      <div>
        {{range .Strings}}
        <span class="mstr">{{.Content}}</span>
        {{end}}
      </div>
    </div>

  </div>
</div>
{{end}}

</div>

<script>

// expand/collapse
function expandAll() {
  document.querySelectorAll('.finding').forEach(f => f.classList.add('expanded'));
}

function collapseAll() {
  document.querySelectorAll('.finding').forEach(f => f.classList.remove('expanded'));
}

// click expand
document.querySelectorAll('.finding-header').forEach(h => {
  h.onclick = () => {
    h.parentElement.classList.toggle('expanded');
  };
});

// filter
document.querySelectorAll('.filter-btn').forEach(btn => {
  btn.onclick = () => {
    document.querySelectorAll('.filter-btn').forEach(b => b.classList.remove('active'));
    btn.classList.add('active');

    const f = btn.dataset.filter;

    document.querySelectorAll('.finding').forEach(el => {
      if (f === 'all' || el.dataset.severity === f) {
        el.style.display = '';
      } else {
        el.style.display = 'none';
      }
    });
  };
});

// search
document.getElementById('searchBox').oninput = function() {
  const q = this.value.toLowerCase();

  document.querySelectorAll('.finding').forEach(el => {
    const text = el.dataset.text.toLowerCase();
    el.style.display = text.includes(q) ? '' : 'none';
  });
};

// copy
document.querySelectorAll('.copy').forEach(el => {
  el.onclick = () => {
    navigator.clipboard.writeText(el.innerText);
  };
});

// score color
document.querySelectorAll('.score-num').forEach(el => {
  const s = parseInt(el.dataset.score);

  if (s >= 80) el.style.color = '#ff4d4f';
  else if (s >= 50) el.style.color = '#faad14';
  else el.style.color = '#1890ff';
});

</script>

</body>
</html>
`
}
