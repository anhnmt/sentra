package report

import (
	"html/template"
	"time"
)

// ReportData holds all data needed to render an HTML report.
type ReportData struct {
	Version     string
	ScanID      string
	GeneratedAt time.Time

	// Device info
	Hostname string
	OS       string
	Arch     string
	IPAddr   string
	User     string

	// Scan info
	ScanStart  time.Time
	ScanEnd    time.Time
	Duration   time.Duration
	Target     string
	RulesDir   string
	Scanned    int64
	Skipped    int64
	MatchCount int64
	ErrorCount int64
	Status     string

	// Settings
	Workers     int
	CommandLine string

	// Findings
	Findings     []Finding
	AlertCount   int
	WarningCount int
	NoticeCount  int
}

// Finding represents a single detection result for the report.
type Finding struct {
	ID          string
	Severity    string // alert, warning, notice
	Score       int
	Module      string
	Target      string
	FileType    string
	RuleName    string
	RuleType    string
	SubScore    int
	Description string
	Author      string
	Date        string
	Class       string
	AttackTags  []string
	Refs        []string
	Strings     []MatchedString

	// Hashes
	MD5    string
	SHA1   string
	SHA256 string
}

// MatchedString represents a YARA string match.
type MatchedString struct {
	Content  string
	Position string
}

// LoadTemplate returns the parsed HTML template.
func LoadTemplate() (*template.Template, error) {
	funcs := template.FuncMap{
		"upper": func(s string) string {
			if s == "" {
				return ""
			}
			return string(s[0]-32) + s[1:]
		},
	}
	tmpl := template.Must(template.New("report").Funcs(funcs).Parse(Template()))
	return tmpl, nil
}

func Template() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Sentra Scan Report</title>
<style>
:root{--bg:#d4d7dc;--bg2:#dde0e5;--bg3:#c8ccd4;--bd:#b8bcc6;--bd2:#a0a6b2;--tx:#1e2530;--tx2:#58626e;--tx3:#0d1018;--ac:#2f5fc4;--ac2:#1e3e8a;--al:#b82218;--al-bg:#e2d0ce;--al-bd:#c09490;--wn:#9a5800;--wn-bg:#e2d8c0;--wn-bd:#c4ae6a;--nt:#176636;--nt-bg:#c4ddd0;--nt-bd:#78b898;--tag:#c4c8d0;--hash:#1e3e8a;--r:4px;--r2:3px;--mono:'Cascadia Code','Consolas','Menlo','Monaco',monospace;--sans:'Segoe UI',system-ui,-apple-system,sans-serif}
*{box-sizing:border-box;margin:0;padding:0}
body{font-family:var(--sans);background:var(--bg);color:var(--tx);min-height:100vh;font-size:15px}
.container{max-width:1400px;margin:0 auto;padding:24px 32px}
.header{display:flex;align-items:center;gap:12px;margin-bottom:20px;border-bottom:1px solid var(--bd);padding-bottom:14px}
.logo{font-size:22px;font-weight:700;letter-spacing:1.5px;color:var(--tx3)}
.logo span{color:var(--ac)}
.version{font-family:var(--mono);font-size:14px;color:var(--ac);background:rgba(47,95,196,.1);border:1px solid rgba(47,95,196,.3);padding:2px 8px;border-radius:var(--r2)}
.header-right{margin-left:auto;font-family:var(--mono);font-size:14px;color:var(--tx2)}
.meta-grid{display:grid;grid-template-columns:repeat(4,1fr);gap:8px;margin-bottom:8px}
.card{background:var(--bg2);border:1px solid var(--bd);border-radius:var(--r);padding:11px 14px}
.card-label{font-size:14px;letter-spacing:1.2px;color:var(--tx2);text-transform:uppercase;margin-bottom:4px;font-family:var(--mono)}
.card-value{font-family:var(--mono);font-size:14px;color:var(--tx3)}
.card-value .tag{display:inline-block;background:var(--tag);border:1px solid var(--bd);padding:1px 7px;border-radius:var(--r2);font-size:14px}
.ok{color:var(--nt)!important;font-weight:600}
.bad{color:var(--al)!important;font-weight:600}
.cmd-card{background:var(--bg2);border:1px solid var(--bd);border-radius:var(--r);padding:11px 14px;margin-bottom:18px}
.cmd-value{font-family:var(--mono);font-size:14px;color:var(--ac2);background:var(--bg3);border:1px solid var(--bd);padding:6px 11px;border-radius:var(--r);margin-top:5px;display:inline-block;white-space:pre-wrap;word-break:break-all}
.filter-bar{display:flex;flex-direction:column;gap:7px;position:sticky;top:0;z-index:100;background:var(--bg);border-bottom:1px solid var(--bd);padding:9px 20px;margin:0 -20px 12px}
.filter-row{display:flex;align-items:center;gap:7px;flex-wrap:wrap}
.modules-row{padding-top:6px;border-top:1px solid var(--bd)}
.mod-row-label{font-family:var(--mono);font-size:14px;letter-spacing:1.2px;text-transform:uppercase;color:var(--tx2);margin-right:4px;white-space:nowrap}
.tab,.expand-btn{display:inline-flex;align-items:center;gap:5px;border:1px solid var(--bd);border-radius:var(--r);background:var(--bg2);color:var(--tx2);cursor:pointer;font-family:var(--sans);transition:border-color .12s,color .12s,background .12s}
.tab{padding:4px 12px;font-size:14px;font-weight:500}
.expand-btn{padding:4px 10px;font-size:14px;font-family:var(--mono)}
.tab:hover,.expand-btn:hover{border-color:var(--ac);color:var(--ac)}
.tab.active{background:var(--tx3);border-color:var(--tx3);color:#fff;font-weight:600}
.tab.alert-tab.active{background:var(--al);border-color:var(--al);color:#fff}
.tab.warn-tab.active{background:var(--wn);border-color:var(--wn);color:#fff}
.tab-dot{width:6px;height:6px;border-radius:50%;display:inline-block;flex-shrink:0}
.tab-count{min-width:18px;text-align:center;background:rgba(0,0,0,.08);border-radius:8px;padding:0 5px;font-size:14px;font-family:var(--mono)}
.tab.active .tab-count{background:rgba(255,255,255,.2)}
.expand-btns{display:flex;gap:4px}
.mod-chip{display:inline-flex;align-items:center;gap:5px;border:1px solid var(--bd);border-radius:20px;padding:3px 10px;font-size:14px;cursor:pointer;transition:border-color .12s,color .12s,background .12s;background:var(--bg2);color:var(--tx2);white-space:nowrap;font-family:var(--mono)}
.mod-chip:hover,.mod-chip.active-mod{border-color:var(--ac);color:var(--ac);background:rgba(47,95,196,.08)}
.mod-chip.active-mod{font-weight:600}
.mod-count{font-size:14px;background:var(--tag);padding:1px 5px;border-radius:8px}
.finding{background:var(--bg2);border:1px solid var(--bd);border-radius:var(--r);margin-bottom:6px;overflow:hidden;transition:box-shadow .15s;animation:fi .2s ease both}
.finding:hover{box-shadow:0 2px 6px rgba(0,0,0,.1)}
@keyframes fi{from{opacity:0;transform:translateY(4px)}to{opacity:1;transform:translateY(0)}}
.finding-header{display:flex;align-items:center;gap:10px;padding:10px 14px;border-left:3px solid transparent;cursor:pointer}
.alert-card .finding-header{border-left-color:var(--al);background:var(--al-bg)}
.warn-card .finding-header{border-left-color:var(--wn);background:var(--wn-bg)}
.notice-card .finding-header{border-left-color:var(--nt);background:var(--nt-bg)}
.severity-badge{font-weight:700;font-size:14px;padding:2px 7px;border-radius:var(--r2);text-transform:uppercase;letter-spacing:1px;flex-shrink:0;color:#fff}
.severity-badge.alert{background:var(--al)}
.severity-badge.warn{background:var(--wn)}
.severity-badge.notice{background:var(--nt)}
.score-num{font-family:var(--mono);font-size:20px;font-weight:700;min-width:40px;flex-shrink:0;line-height:1}
.score-num.alert{color:var(--al)}
.score-num.warn{color:var(--wn)}
.score-num.notice{color:var(--nt)}
.header-mid{flex:1;min-width:0;display:flex;flex-direction:column;gap:3px}
.file-path{font-family:var(--mono);font-size:14px;color:var(--hash);overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.header-tags{display:flex;gap:4px;flex-wrap:wrap}
.htag{font-family:var(--mono);font-size:14px;padding:1px 6px;border-radius:var(--r2);border:1px solid var(--bd);color:var(--tx2);background:var(--tag);white-space:nowrap}
.htag.module{border-color:rgba(47,95,196,.35);color:var(--ac2);background:rgba(47,95,196,.1)}
.htag.sigtype{border-color:rgba(154,88,0,.3);color:var(--wn);background:rgba(154,88,0,.08)}
.htag.atk{border-color:rgba(90,47,150,.3);color:#5a2f96;background:rgba(90,47,150,.08)}
.file-type-badge{font-family:var(--mono);font-size:14px;color:var(--tx2);border:1px solid var(--bd);padding:2px 7px;border-radius:var(--r2);flex-shrink:0;background:var(--tag)}
.finding-body{padding:14px;display:none;border-top:1px solid var(--bd);background:var(--bg3)}
.finding.expanded .finding-body{display:block}
.alert-card.expanded .finding-body{background:var(--al-bg);border-top-color:var(--al-bd)}
.warn-card.expanded .finding-body{background:var(--wn-bg);border-top-color:var(--wn-bd)}
.notice-card.expanded .finding-body{background:var(--nt-bg);border-top-color:var(--nt-bd)}
.body-section{margin-bottom:14px}
.body-section:last-child{margin-bottom:0}
.section-title{font-size:14px;letter-spacing:1.5px;text-transform:uppercase;color:var(--tx2);font-family:var(--mono);margin-bottom:7px;border-bottom:1px solid var(--bd);padding-bottom:4px}
.kv-grid{display:grid;grid-template-columns:140px 1fr;gap:2px 10px;font-family:var(--mono);font-size:14px}
.kv-key{color:var(--tx2);padding:2px 0;white-space:nowrap}
.kv-val{color:var(--tx3);padding:2px 0;word-break:break-all}
.kv-val.hash,.kv-val.path{color:var(--hash);word-break:break-all}
.kv-val.cmd{color:var(--wn);white-space:pre-wrap;word-break:break-all}
.reason-box{background:var(--bg2);border:1px solid var(--bd);border-radius:var(--r);padding:11px 13px;margin-bottom:7px}
.reason-header{display:flex;align-items:center;gap:8px;flex-wrap:wrap;margin-bottom:5px}
.reason-name{font-weight:600;font-size:14px;color:var(--tx3);font-family:var(--mono)}
.subscore-tag{font-family:var(--mono);font-size:14px;color:var(--tx2);background:var(--tag);border:1px solid var(--bd);padding:1px 7px;border-radius:var(--r2)}
.sigtype-tag{font-family:var(--mono);font-size:14px;padding:2px 7px;border-radius:var(--r2);border:1px solid rgba(154,88,0,.3);color:var(--wn);background:rgba(154,88,0,.08)}
.reason-desc{font-size:14px;color:var(--tx2);margin-bottom:4px;line-height:1.5}
.reason-meta{font-size:14px;color:var(--tx2);margin-bottom:5px;display:flex;gap:14px;flex-wrap:wrap}
.reason-meta strong{color:var(--tx)}
.ref-link{font-family:var(--mono);font-size:14px;color:var(--ac);text-decoration:underline;text-underline-offset:2px;display:block;margin-bottom:7px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.atk-tags{display:flex;gap:4px;flex-wrap:wrap;margin-bottom:7px}
.atk-tag{font-family:var(--mono);font-size:14px;padding:2px 7px;border-radius:var(--r2);border:1px solid rgba(90,47,150,.3);color:#5a2f96;background:rgba(90,47,150,.08)}
.matched-strings{display:flex;flex-wrap:wrap;gap:4px}
.mstr{font-family:var(--mono);font-size:14px;background:rgba(47,95,196,.09);border:1px solid rgba(47,95,196,.28);color:var(--ac2);padding:2px 9px;border-radius:var(--r2)}
.context-box{background:var(--bg2);border:1px solid var(--bd);border-radius:var(--r);padding:9px 11px;font-family:var(--mono);font-size:14px;color:var(--tx);white-space:pre-wrap;word-break:break-all;max-height:140px;overflow-y:auto;line-height:1.6;margin-top:7px}
.context-box mark{background:#d8cc50;color:#4a3000;border-radius:2px;padding:0 2px}
.showing-count{font-family:var(--mono);font-size:12px;color:var(--tx2);white-space:nowrap}
.showing-count span{color:var(--ac);font-weight:600}
.empty-state{text-align:center;padding:40px 20px;color:var(--tx2);font-family:var(--mono);font-size:13px;background:var(--bg2);border:1px dashed var(--bd);border-radius:var(--r);margin-top:8px}
.footer{text-align:center;padding:22px;color:var(--tx2);font-family:var(--mono);font-size:14px;border-top:1px solid var(--bd);margin-top:24px}
::-webkit-scrollbar{width:5px;height:5px}
::-webkit-scrollbar-track{background:var(--bg3)}
::-webkit-scrollbar-thumb{background:var(--bd2);border-radius:3px}
@media(max-width:1024px){.container{padding:16px 16px}.meta-grid{grid-template-columns:repeat(2,1fr)}.filter-bar{padding:8px 16px;margin:0 -16px 12px}.kv-grid{grid-template-columns:120px 1fr}}
@media(max-width:768px){.container{padding:12px}.meta-grid{grid-template-columns:repeat(2,1fr)}.filter-bar{padding:8px 12px;margin:0 -12px 10px}.filter-row{gap:5px}.tab{padding:4px 9px;font-size:11px}.mod-chip{padding:3px 8px;font-size:11px}.expand-btns{display:none}.kv-grid{grid-template-columns:100px 1fr}.score-num{font-size:17px;min-width:32px}.finding-header{padding:9px 11px;gap:7px}.file-type-badge{display:none}.header-right{display:none}}
@media(max-width:480px){.meta-grid{grid-template-columns:1fr 1fr}.header-tags{display:none}.mod-row-label{display:none}}
</style>
</head>
<body>
<div class="container">

<!-- HEADER -->
<div class="header">
  <div class="logo">SENTRA<span> //</span></div>
  <div class="version">{{.Version}}</div>
  <div class="header-right">Scan ID: <strong class="ac">{{.ScanID}}</strong> &nbsp;·&nbsp; Generated: {{.GeneratedAt.Format "2006-01-02 15:04:05 UTC"}}</div>
</div>

<!-- SCAN INFO -->
<div class="meta-grid">
  <div class="card"><div class="card-label">Hostname</div><div class="card-value"><span class="tag">{{.Hostname}}</span></div></div>
  <div class="card"><div class="card-label">IP Address</div><div class="card-value">{{.IPAddr}}</div></div>
  <div class="card"><div class="card-label">Platform</div><div class="card-value">{{.OS}} {{.Arch}}</div></div>
  <div class="card"><div class="card-label">Run As User</div><div class="card-value">{{.User}}</div></div>
</div>
<div class="meta-grid">
  <div class="card"><div class="card-label">Scan Start</div><div class="card-value">{{.ScanStart.Format "2006-01-02 15:04:05 UTC"}}</div></div>
  <div class="card"><div class="card-label">Scan End</div><div class="card-value">{{.ScanEnd.Format "2006-01-02 15:04:05 UTC"}}</div></div>
  <div class="card"><div class="card-label">Duration</div><div class="card-value">{{.Duration}}</div></div>
  <div class="card"><div class="card-label">Status</div><div class="card-value">{{.Status}}</div></div>
</div>
<div class="meta-grid">
  <div class="card"><div class="card-label">Signature DB</div><div class="card-value">{{.RulesDir}}</div></div>
  <div class="card"><div class="card-label">Files Scanned</div><div class="card-value">{{.Scanned}}</div></div>
  <div class="card"><div class="card-label">Files Skipped</div><div class="card-value">{{.Skipped}}</div></div>
  <div class="card"><div class="card-label">Errors</div><div class="card-value">{{.ErrorCount}}</div></div>
</div>

<!-- CMD -->
<div class="cmd-card">
  <div class="card-label">Command Line</div>
  <div class="cmd-value">{{.CommandLine}}</div>
</div>

<!-- FILTER BAR -->
<div class="filter-bar">
  <div class="filter-row">
    <button class="tab active" onclick="filterFindings('all',this)">All <span class="tab-count" id="cnt-all">{{len .Findings}}</span></button>
    <button class="tab alert-tab" onclick="filterFindings('alert',this)"><span class="tab-dot"></span>Alert <span class="tab-count" id="cnt-alert">{{.AlertCount}}</span></button>
    <button class="tab warn-tab"  onclick="filterFindings('warning',this)"><span class="tab-dot"></span>Warning <span class="tab-count" id="cnt-warn">{{.WarningCount}}</span></button>
    <button class="tab"           onclick="filterFindings('notice',this)"><span class="tab-dot"></span>Notice <span class="tab-count" id="cnt-notice">{{.NoticeCount}}</span></button>
    <div class="expand-btns">
      <button class="expand-btn" onclick="expandAll()">⊞ Expand All</button>
      <button class="expand-btn" onclick="collapseAll()">⊟ Collapse All</button>
    </div>
    <span id="showing-count" class="showing-count"></span>
  </div>
</div>

<!-- FINDINGS -->
<div id="findings">
{{range .Findings}}
<!-- {{.RuleName}} -->
<div class="finding {{.Severity}}-card" data-severity="{{.Severity}}" data-module="{{.Module}}"
     data-path="{{.Target}}" data-score="{{.Score}}">
  <div class="finding-header" onclick="toggleFinding(this)">
    <span class="severity-badge {{.Severity}}">{{upper .Severity}}</span>
    <span class="score-num {{.Severity}}">{{.Score}}</span>
    <div class="header-mid">
      <span class="file-path" title="{{.Target}}">{{.Target}}</span>
      <div class="header-tags">
        <span class="htag module">{{.Module}}</span>
        <span class="htag sigtype">{{.RuleType}}</span>
        {{range .AttackTags}}<span class="htag atk">{{.}}</span>{{end}}
      </div>
    </div>
    <span class="file-type-badge">{{.FileType}}</span>
  </div>
  <div class="finding-body">
    <div class="body-section">
      <div class="section-title">Match Details</div>
      <div class="kv-grid">
        <span class="kv-key">Rule</span><span class="kv-val">{{.RuleName}}</span>
        <span class="kv-key">Description</span><span class="kv-val">{{.Description}}</span>
        {{if .Author}}<span class="kv-key">Author</span><span class="kv-val">{{.Author}}</span>{{end}}
        {{if .Date}}<span class="kv-key">Date</span><span class="kv-val">{{.Date}}</span>{{end}}
        {{if .Class}}<span class="kv-key">Class</span><span class="kv-val">{{.Class}}</span>{{end}}
        {{if .MD5}}<span class="kv-key">MD5</span><span class="kv-val hash">{{.MD5}}</span>{{end}}
        {{if .SHA256}}<span class="kv-key">SHA256</span><span class="kv-val hash">{{.SHA256}}</span>{{end}}
      </div>
    </div>
    {{if .Strings}}
    <div class="body-section">
      <div class="section-title">Matched Strings</div>
      <div class="matched-strings">
        {{range .Strings}}<span class="mstr">{{.Content}} @ {{.Position}}</span>{{end}}
      </div>
    </div>
    {{end}}
    {{if .Refs}}
    <div class="body-section">
      <div class="section-title">References</div>
      {{range .Refs}}<a class="ref-link" href="#">{{.}}</a>{{end}}
    </div>
    {{end}}
  </div>
</div>
{{end}}
{{if not .Findings}}
<div class="empty-state">No findings detected</div>
{{end}}
</div>

<!-- FOOTER -->
<div class="footer">
  Generated by Sentra v{{.Version}} &nbsp;·&nbsp; {{.GeneratedAt.Format "2006-01-02 15:04:05 UTC"}}
</div>

</div>

<script>
function toggleFinding(el) {
  el.closest('.finding').classList.toggle('expanded');
}
function filterFindings(sev, btn) {
  document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
  btn.classList.add('active');
  document.querySelectorAll('.finding').forEach(f => {
    f.style.display = (sev === 'all' || f.dataset.severity === sev) ? 'block' : 'none';
  });
}
function expandAll() {
  document.querySelectorAll('.finding').forEach(f => f.classList.add('expanded'));
}
function collapseAll() {
  document.querySelectorAll('.finding').forEach(f => f.classList.remove('expanded'));
}
</script>
</body>
</html>`
}

var tmpl = template.Must(template.New("report").Funcs(template.FuncMap{
	"upper": func(s string) string {
		if s == "" {
			return ""
		}
		return string(s[0]-32) + s[1:]
	},
}).Parse(Template()))
