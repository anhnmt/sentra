package report

func Template() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Sentra Scan Report</title>
<style>:root{--bg:#d4d7dc;--bg2:#dde0e5;--bg3:#c8ccd4;--bd:#b8bcc6;--bd2:#a0a6b2;--tx:#1e2530;--tx2:#58626e;--tx3:#0d1018;--ac:#2f5fc4;--ac2:#1e3e8a;--al:#b82218;--al-bg:#e2d0ce;--al-bd:#c09490;--wn:#9a5800;--wn-bg:#e2d8c0;--wn-bd:#c4ae6a;--nt:#176636;--nt-bg:#c4ddd0;--nt-bd:#78b898;--tag:#c4c8d0;--hash:#1e3e8a;--r:4px;--r2:3px;--mono:'Cascadia Code','Consolas','Menlo','Monaco',monospace;--sans:'Segoe UI',system-ui,-apple-system,sans-serif}*{box-sizing:border-box;margin:0;padding:0}body{font-family:var(--sans);background:var(--bg);color:var(--tx);min-height:100vh;font-size:15px}.container{max-width:1400px;margin:0 auto;padding:24px 32px}.header{display:flex;align-items:center;gap:12px;margin-bottom:20px;border-bottom:1px solid var(--bd);padding-bottom:14px}.logo{font-size:22px;font-weight:700;letter-spacing:1.5px;color:var(--tx3)}.logo span{color:var(--ac)}.version{font-family:var(--mono);font-size:14px;color:var(--ac);background:rgba(47,95,196,.1);border:1px solid rgba(47,95,196,.3);padding:2px 8px;border-radius:var(--r2)}.header-right{margin-left:auto;font-family:var(--mono);font-size:14px;color:var(--tx2)}.meta-grid{display:grid;grid-template-columns:repeat(4,1fr);gap:8px;margin-bottom:8px}.card{background:var(--bg2);border:1px solid var(--bd);border-radius:var(--r);padding:11px 14px}.card-label{font-size:14px;letter-spacing:1.2px;color:var(--tx2);text-transform:uppercase;margin-bottom:4px;font-family:var(--mono)}.card-value{font-family:var(--mono);font-size:14px;color:var(--tx3)}.card-value .tag{display:inline-block;background:var(--tag);border:1px solid var(--bd);padding:1px 7px;border-radius:var(--r2);font-size:14px}.ok{color:var(--nt)!important;font-weight:600}.bad{color:var(--al)!important;font-weight:600}.cmd-card{background:var(--bg2);border:1px solid var(--bd);border-radius:var(--r);padding:11px 14px;margin-bottom:18px}.cmd-value{font-family:var(--mono);font-size:14px;color:var(--ac2);background:var(--bg3);border:1px solid var(--bd);padding:6px 11px;border-radius:var(--r);margin-top:5px;display:inline-block;white-space:pre-wrap;word-break:break-all}.filter-bar{display:flex;flex-direction:column;gap:7px;position:sticky;top:0;z-index:100;background:var(--bg);border-bottom:1px solid var(--bd);padding:9px 20px;margin:0 -20px 12px}.filter-row{display:flex;align-items:center;gap:7px;flex-wrap:wrap}.modules-row{padding-top:6px;border-top:1px solid var(--bd)}.mod-row-label{font-family:var(--mono);font-size:14px;letter-spacing:1.2px;text-transform:uppercase;color:var(--tx2);margin-right:4px;white-space:nowrap}.tab,.expand-btn{display:inline-flex;align-items:center;gap:5px;border:1px solid var(--bd);border-radius:var(--r);background:var(--bg2);color:var(--tx2);cursor:pointer;font-family:var(--sans);transition:border-color .12s,color .12s,background .12s}.tab{padding:4px 12px;font-size:14px;font-weight:500}.expand-btn{padding:4px 10px;font-size:14px;font-family:var(--mono)}.tab:hover,.expand-btn:hover{border-color:var(--ac);color:var(--ac)}.tab.active{background:var(--tx3);border-color:var(--tx3);color:#fff;font-weight:600}.tab.alert-tab.active{background:var(--al);border-color:var(--al);color:#fff}.tab.warn-tab.active{background:var(--wn);border-color:var(--wn);color:#fff}.tab-dot{width:6px;height:6px;border-radius:50%;display:inline-block;flex-shrink:0}.td-al{background:var(--al)}.td-wn{background:var(--wn)}.td-nt{background:var(--nt)}.tab-count{min-width:18px;text-align:center;background:rgba(0,0,0,.08);border-radius:8px;padding:0 5px;font-size:14px;font-family:var(--mono)}.tab.active .tab-count{background:rgba(255,255,255,.2)}.expand-btns{display:flex;gap:4px}.mod-chip{display:inline-flex;align-items:center;gap:5px;border:1px solid var(--bd);border-radius:20px;padding:3px 10px;font-size:14px;cursor:pointer;transition:border-color .12s,color .12s,background .12s;background:var(--bg2);color:var(--tx2);white-space:nowrap;font-family:var(--mono)}.mod-chip:hover,.mod-chip.active-mod{border-color:var(--ac);color:var(--ac);background:rgba(47,95,196,.08)}.mod-chip.active-mod{font-weight:600}.mod-count{font-size:14px;background:var(--tag);padding:1px 5px;border-radius:8px}.active-filters{background:var(--bg2);border:1px solid var(--bd);border-radius:var(--r);padding:7px 12px;margin-bottom:7px;display:flex;align-items:center;gap:7px;flex-wrap:wrap}.af-label{font-size:14px;color:var(--tx2);text-transform:uppercase;letter-spacing:1px;font-family:var(--mono);display:flex;align-items:center;gap:5px;white-space:nowrap}.af-dot{width:5px;height:5px;background:var(--ac);border-radius:50%}.filter-chip,.exclude-chip{border:1px solid var(--bd);padding:2px 8px;border-radius:var(--r2);font-family:var(--mono);font-size:14px;display:inline-flex;align-items:center;gap:4px}.filter-chip{background:var(--tag);color:var(--tx)}.filter-chip.add-chip,.exclude-chip.add-chip{background:transparent;border-style:dashed;color:var(--tx2)}.filter-chip .x,.exclude-chip .x{cursor:pointer;font-size:14px;color:var(--tx2)}.exclude-chip{background:var(--al-bg);border-color:var(--al-bd);color:var(--al)}.exclude-panel{border-color:var(--al-bd)}.ex-prefix{font-weight:700;margin-right:2px}.kw-input{background:transparent;border:none;outline:none;color:var(--tx);font-family:var(--mono);font-size:14px;width:120px}.kw-input::placeholder{color:var(--tx2)}.af-actions{margin-left:auto}.af-btn{background:transparent;border:1px solid var(--bd);color:var(--tx2);padding:2px 8px;border-radius:var(--r2);font-size:14px;cursor:pointer;font-family:var(--sans);font-weight:500;transition:border-color .12s,color .12s}.af-btn:hover{border-color:var(--bd2);color:var(--tx)}.finding{background:var(--bg2);border:1px solid var(--bd);border-radius:var(--r);margin-bottom:6px;overflow:hidden;transition:box-shadow .15s;animation:fi .2s ease both}.finding:hover{box-shadow:0 2px 6px rgba(0,0,0,.1)}@keyframes fi{from{opacity:0;transform:translateY(4px)}to{opacity:1;transform:translateY(0)}}.finding-header{display:flex;align-items:center;gap:10px;padding:10px 14px;border-left:3px solid transparent;cursor:pointer}.alert-card .finding-header,.high-card .finding-header{border-left-color:var(--al);background:var(--al-bg)}.warn-card .finding-header,.warning-card .finding-header{border-left-color:var(--wn);background:var(--wn-bg)}.notice-card .finding-header{border-left-color:var(--nt);background:var(--nt-bg)}.severity-badge{font-weight:700;font-size:14px;padding:2px 7px;border-radius:var(--r2);text-transform:uppercase;letter-spacing:1px;flex-shrink:0;color:#fff}.severity-badge.alert,.severity-badge.high{background:var(--al)}.severity-badge.warn,.severity-badge.warning{background:var(--wn)}.severity-badge.notice{background:var(--nt)}.score-num{font-family:var(--mono);font-size:20px;font-weight:700;min-width:40px;flex-shrink:0;line-height:1}.score-num.alert,.score-num.high{color:var(--al)}.score-num.warn,.score-num.warning{color:var(--wn)}.score-num.notice{color:var(--nt)}.header-mid{flex:1;min-width:0;display:flex;flex-direction:column;gap:3px}.file-path{font-family:var(--mono);font-size:14px;color:var(--hash);overflow:hidden;text-overflow:ellipsis;white-space:nowrap}.header-tags{display:flex;gap:4px;flex-wrap:wrap}.htag{font-family:var(--mono);font-size:14px;padding:1px 6px;border-radius:var(--r2);border:1px solid var(--bd);color:var(--tx2);background:var(--tag);white-space:nowrap}.htag.module{border-color:rgba(47,95,196,.35);color:var(--ac2);background:rgba(47,95,196,.1)}.htag.sigtype{border-color:rgba(154,88,0,.3);color:var(--wn);background:rgba(154,88,0,.08)}.htag.atk{border-color:rgba(90,47,150,.3);color:#5a2f96;background:rgba(90,47,150,.08)}.file-type-badge{font-family:var(--mono);font-size:14px;color:var(--tx2);border:1px solid var(--bd);padding:2px 7px;border-radius:var(--r2);flex-shrink:0;background:var(--tag)}.finding-body{padding:14px;display:none;border-top:1px solid var(--bd);background:var(--bg3)}.finding.expanded .finding-body{display:block}.alert-card.expanded .finding-body,.high-card.expanded .finding-body{background:var(--al-bg);border-top-color:var(--al-bd)}.warn-card.expanded .finding-body,.warning-card.expanded .finding-body{background:var(--wn-bg);border-top-color:var(--wn-bd)}.notice-card.expanded .finding-body{background:var(--nt-bg);border-top-color:var(--nt-bd)}.body-section{margin-bottom:14px}.body-section:last-child{margin-bottom:0}.section-title{font-size:14px;letter-spacing:1.5px;text-transform:uppercase;color:var(--tx2);font-family:var(--mono);margin-bottom:7px;border-bottom:1px solid var(--bd);padding-bottom:4px}.kv-grid{display:grid;grid-template-columns:140px 1fr;gap:2px 10px;font-family:var(--mono);font-size:14px}.kv-key{color:var(--tx2);padding:2px 0;white-space:nowrap}.kv-val{color:var(--tx3);padding:2px 0;word-break:break-all}.kv-val.hash,.kv-val.path{color:var(--hash);word-break:break-all}.kv-val.cmd{color:var(--wn);white-space:pre-wrap;word-break:break-all}.reason-box{background:var(--bg2);border:1px solid var(--bd);border-radius:var(--r);padding:11px 13px;margin-bottom:7px}.reason-header{display:flex;align-items:center;gap:8px;flex-wrap:wrap;margin-bottom:5px}.reason-name{font-weight:600;font-size:14px;color:var(--tx3);font-family:var(--mono)}.subscore-tag{font-family:var(--mono);font-size:14px;color:var(--tx2);background:var(--tag);border:1px solid var(--bd);padding:1px 7px;border-radius:var(--r2)}.sigtype-tag{font-family:var(--mono);font-size:14px;padding:2px 7px;border-radius:var(--r2);border:1px solid rgba(154,88,0,.3);color:var(--wn);background:rgba(154,88,0,.08)}.reason-desc{font-size:14px;color:var(--tx2);margin-bottom:4px;line-height:1.5}.reason-meta{font-size:14px;color:var(--tx2);margin-bottom:5px;display:flex;gap:14px;flex-wrap:wrap}.reason-meta strong{color:var(--tx)}.ref-link{font-family:var(--mono);font-size:14px;color:var(--ac);text-decoration:underline;text-underline-offset:2px;display:block;margin-bottom:7px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}.atk-tags{display:flex;gap:4px;flex-wrap:wrap;margin-bottom:7px}.atk-tag{font-family:var(--mono);font-size:14px;padding:2px 7px;border-radius:var(--r2);border:1px solid rgba(90,47,150,.3);color:#5a2f96;background:rgba(90,47,150,.08)}.matched-strings{display:flex;flex-wrap:wrap;gap:4px}.mstr{font-family:var(--mono);font-size:14px;background:rgba(47,95,196,.09);border:1px solid rgba(47,95,196,.28);color:var(--ac2);padding:2px 9px;border-radius:var(--r2)}.context-box{background:var(--bg2);border:1px solid var(--bd);border-radius:var(--r);padding:9px 11px;font-family:var(--mono);font-size:14px;color:var(--tx);white-space:pre-wrap;word-break:break-all;max-height:140px;overflow-y:auto;line-height:1.6;margin-top:7px}.context-box mark{background:#d8cc50;color:#4a3000;border-radius:2px;padding:0 2px}mark.sh{background:#f5d84a;color:#2a1a00;border-radius:2px;padding:0 1px;font-style:normal}.showing-count{font-family:var(--mono);font-size:12px;color:var(--tx2);white-space:nowrap}.showing-count span{color:var(--ac);font-weight:600}.empty-state{text-align:center;padding:40px 20px;color:var(--tx2);font-family:var(--mono);font-size:13px;background:var(--bg2);border:1px dashed var(--bd);border-radius:var(--r);margin-top:8px}#sel-tip{position:fixed;display:none;z-index:9999;background:var(--tx3);border:1px solid var(--bd2);border-radius:var(--r);padding:3px 4px;box-shadow:0 4px 14px rgba(0,0,0,.18);display:none;align-items:center;gap:2px;font-family:var(--mono);font-size:11px;white-space:nowrap}#sel-tip button{background:transparent;border:none;cursor:pointer;padding:3px 8px;border-radius:var(--r2);font-family:var(--mono);font-size:11px;transition:background .1s}#sel-tip .tip-in{color:#6ef08a}#sel-tip .tip-in:hover{background:rgba(110,240,138,.15)}#sel-tip .tip-out{color:#f08a6e}#sel-tip .tip-out:hover{background:rgba(240,138,110,.15)}#sel-tip .tip-div{width:1px;height:14px;background:var(--bd)}.vt-hash{color:var(--hash);text-decoration:underline;text-decoration-style:dotted;text-underline-offset:2px;word-break:break-all}.vt-hash:hover{color:var(--ac)}.footer{text-align:center;padding:22px;color:var(--tx2);font-family:var(--mono);font-size:14px;border-top:1px solid var(--bd);margin-top:24px}::-webkit-scrollbar{width:5px;height:5px}::-webkit-scrollbar-track{background:var(--bg3)}::-webkit-scrollbar-thumb{background:var(--bd2);border-radius:3px}@media(max-width:1024px){.container{padding:16px 16px}.meta-grid{grid-template-columns:repeat(2,1fr)}.filter-bar{padding:8px 16px;margin:0 -16px 12px}.kv-grid{grid-template-columns:120px 1fr}}@media(max-width:768px){.container{padding:12px}.meta-grid{grid-template-columns:repeat(2,1fr)}.filter-bar{padding:8px 12px;margin:0 -12px 10px}.filter-row{gap:5px}.tab{padding:4px 9px;font-size:11px}.mod-chip{padding:3px 8px;font-size:11px}.expand-btns{display:none}.kv-grid{grid-template-columns:100px 1fr}.score-num{font-size:17px;min-width:32px}.finding-header{padding:9px 11px;gap:7px}.file-type-badge{display:none}.header-right{display:none}}@media(max-width:480px){.meta-grid{grid-template-columns:1fr 1fr}.header-tags{display:none}.mod-row-label{display:none}}</style>

</head>
<body>
<div class="container">

<!-- ══ HEADER ══════════════════════════════════════════════════ -->
<div class="header">
  <div class="logo">SENTRA<span> //</span></div>
  <div class="version">v{{.Version}}</div>
  <div class="header-right">Scan ID: <strong class="ac">{{.ScanID}}</strong> &nbsp;·&nbsp; Generated: {{.GeneratedAt.Format "2006-01-02 15:04:05 MST"}}</div>
</div>

<!-- ══ SCAN INFO ════════════════════════════════════════════════ -->
<div class="meta-grid">
  <div class="card"><div class="card-label">Hostname</div><div class="card-value"><span class="tag">{{.Hostname}}</span></div></div>
  <div class="card"><div class="card-label">IP Address</div><div class="card-value">{{.IPAddr}}</div></div>
  <div class="card"><div class="card-label">Platform</div><div class="card-value">{{.OS}}</div></div>
  <div class="card"><div class="card-label">Run As User</div><div class="card-value">{{.User}}</div></div>
</div>
<div class="meta-grid">
  <div class="card"><div class="card-label">Scan Start</div><div class="card-value">{{.ScanStart.Format "2006-01-02 15:04:05 MST"}}</div></div>
  <div class="card"><div class="card-label">Scan End</div><div class="card-value">{{.ScanEnd.Format "2006-01-02 15:04:05 MST"}}</div></div>
  <div class="card"><div class="card-label">Duration</div><div class="card-value">{{.Duration}}</div></div>
  
</div>
<div class="meta-grid cols2">
  <div class="card"><div class="card-label">Signature DB</div><div class="card-value">{{.RulesDir}}</div></div>
  
</div>
<div class="meta-grid cols3">
  <div class="card"><div class="card-label">Thresholds</div><div class="card-value" class="cv-sm">Alert: 80 · Warning: 60 · Notice: 40</div></div>
  <div class="card"><div class="card-label">Scan Settings</div><div class="card-value" class="cv-sm">Workers: {{.Workers}}</div></div>
  
</div>

<!-- CMD -->
<div class="cmd-card">
  <div class="card-label">Command Line</div>
  <div class="cmd-value">{{.CommandLine}}</div>
</div>

<!-- ══ STICKY FILTER BAR ════════════════════════════════════════ -->
<div class="filter-bar">
  <div class="filter-row">
    <button class="tab active" onclick="filterFindings('all',this)">All <span class="tab-count" id="cnt-all">{{len .Findings}}</span></button>
    <button class="tab alert-tab" onclick="filterFindings('alert',this)"><span class="tab-dot td-al"></span>Alert <span class="tab-count" id="cnt-alert">{{.AlertCount}}</span></button>
    <button class="tab warn-tab"  onclick="filterFindings('warning',this)"><span class="tab-dot td-wn"></span>Warning <span class="tab-count" id="cnt-warn">{{.WarningCount}}</span></button>
    <button class="tab"           onclick="filterFindings('notice',this)"><span class="tab-dot td-nt"></span>Notice <span class="tab-count" id="cnt-notice">{{.NoticeCount}}</span></button>
    <div class="expand-btns">
      <button class="expand-btn" onclick="expandAll()">⊞ Expand All</button>
      <button class="expand-btn" onclick="collapseAll()">⊟ Collapse All</button>
    </div>
    <span id="showing-count" class="showing-count"></span>
  </div>
  <div class="filter-row modules-row">
    <span class="mod-row-label">Module</span>
    <span class="mod-chip" data-mod="filescan"       onclick="filterModule('Filescan',this)">Filescan <span class="mod-count">4</span></span>
    <span class="mod-chip" data-mod="eventlog"       onclick="filterModule('Eventlog',this)">Eventlog <span class="mod-count">3</span></span>
    <span class="mod-chip" data-mod="servicecheck"   onclick="filterModule('ServiceCheck',this)">ServiceCheck <span class="mod-count">2</span></span>
    <span class="mod-chip" data-mod="registrychecks" onclick="filterModule('RegistryChecks',this)">RegistryChecks <span class="mod-count">1</span></span>
  </div>
</div>

<!-- ACTIVE FILTERS -->
<div id="active-filters-panel" class="active-filters"></div>
<!-- EXCLUDE FILTERS -->
<div id="exclude-filters-panel" class="active-filters exclude-panel"></div>

<!-- ══════════════════════════════════════════════════════════════
     FINDINGS
═══════════════════════════════════════════════════════════════ -->
<div id="findings">

{{range .Findings}}
<div class="finding {{.Severity}}-card"
     data-severity="{{.Severity}}"
     data-module="{{.Module}}"
     data-path="{{.Target}}"
     data-score="{{.Score}}">
  <div class="finding-header" onclick="toggleFinding(this)">
    <span class="severity-badge {{.Severity}}">{{.Severity}}</span>
    <span class="score-num {{.Severity}}">{{.Score}}</span>
    <div class="header-mid">
      <span class="file-path" title="{{.Target}}">{{.Target}}</span>
      <div class="header-tags">
        <span class="htag module">{{.Module}}</span>
        {{if .RuleType}}<span class="htag sigtype">{{.RuleType}}</span>{{end}}
        {{range .AttackTags}}<span class="htag atk">{{.}}</span>{{end}}
      </div>
    </div>
    <span class="file-type-badge">{{.FileType}}</span>
  </div>
  <div class="finding-body">
    <div class="body-section">
      <div class="section-title">File Info</div>
      <div class="kv-grid">
        <span class="kv-key">File</span><span class="kv-val path">{{.Target}}</span>
        {{if .MD5}}<span class="kv-key">MD5</span><span class="kv-val hash">{{.MD5}}</span>{{end}}
        {{if .SHA1}}<span class="kv-key">SHA1</span><span class="kv-val hash">{{.SHA1}}</span>{{end}}
        {{if .SHA256}}<span class="kv-key">SHA256</span><span class="kv-val hash">{{.SHA256}}</span>{{end}}
      </div>
    </div>
    
    <div class="body-section">
      <div class="section-title">Match Reason</div>
      <div class="reason-box">
        <div class="reason-header">
          <span class="reason-name">{{.RuleName}}</span>
          <span class="subscore-tag">Subscore: {{.SubScore}}</span>
          <span class="sigtype-tag">{{.RuleType}}</span>
        </div>
        {{if .Description}}<div class="reason-desc">{{.Description}}</div>{{end}}
        <div class="reason-meta">
          {{if .Author}}<span><strong>Author:</strong> {{.Author}}</span>{{end}}
          {{if .Date}}<span><strong>Date:</strong> {{.Date}}</span>{{end}}
          {{if .Class}}<span><strong>Class:</strong> {{.Class}}</span>{{end}}
        </div>
        {{if .AttackTags}}<div class="atk-tags">{{range .AttackTags}}<span class="atk-tag">{{.}}</span>{{end}}</div>{{end}}
        {{range .Refs}}<a class="ref-link" href="{{.}}" target="_blank" rel="noopener noreferrer">{{.}}</a>{{end}}
        {{if .Strings}}<div class="matched-strings">{{range .Strings}}<span class="mstr">{{.Content}} @ {{.Position}}</span>{{end}}</div>{{end}}
      </div>
    </div>
    </div>
</div>
</div>
{{end}}
</div><!-- #findings -->

<script>let F='all', S='', M='', inc=[], exc=[];

  const $  = s => document.querySelector(s);
  const $$ = s => document.querySelectorAll(s);
  const el = id => document.getElementById(id);

  const _cache = new Map(); 

  function buildCache() {
    $$('.finding').forEach(f => {

      const body = f.querySelector('.finding-body');
      const header = f.querySelector('.finding-header');
      const bodyTxt  = body   ? body.textContent.toLowerCase()   : '';
      const headerTxt= header ? header.textContent.toLowerCase() : '';
      _cache.set(f, {
        sev:   f.dataset.severity || '',
        path: (f.dataset.path    || '').toLowerCase(),
        mod:  (f.dataset.module  || '').toLowerCase(),
        score: parseInt(f.dataset.score || '0', 10),
        txt:  headerTxt + ' ' + bodyTxt
      });
    });
  }

  function highlightIn(root, q) {
    if (!q) return;
    const re = new RegExp(q.replace(/[.*+?^${}()|[\]\\]/g,'\\$&'), 'gi');

    root.querySelectorAll('a.vt-hash').forEach(a => {
      if (re.test(a.textContent)) {
        const mark = document.createElement('mark');
        mark.className = 'sh';
        a.parentNode.insertBefore(mark, a);
        mark.appendChild(a);
      }
      re.lastIndex = 0;
    });

    const walker = document.createTreeWalker(root, NodeFilter.SHOW_TEXT, {
      acceptNode: n => {
        const p = n.parentNode;
        if (p.classList?.contains('sh'))    return NodeFilter.FILTER_REJECT;
        if (p.nodeName === 'INPUT')          return NodeFilter.FILTER_REJECT;
        if (p.closest?.('.context-box'))     return NodeFilter.FILTER_REJECT;
        if (p.closest?.('.vt-hash'))         return NodeFilter.FILTER_REJECT;
        return NodeFilter.FILTER_ACCEPT;
      }
    });
    const hits = [];
    let node;
    while ((node = walker.nextNode())) { if (re.test(node.nodeValue)) hits.push(node); re.lastIndex=0; }
    hits.forEach(node => {
      const frag = document.createDocumentFragment();
      let last=0, m;
      re.lastIndex=0;
      while ((m = re.exec(node.nodeValue)) !== null) {
        if (m.index > last) frag.appendChild(document.createTextNode(node.nodeValue.slice(last, m.index)));
        const mark = document.createElement('mark');
        mark.className='sh'; mark.textContent=m[0]; frag.appendChild(mark);
        last = m.index + m[0].length;
      }
      if (last < node.nodeValue.length) frag.appendChild(document.createTextNode(node.nodeValue.slice(last)));
      node.parentNode.replaceChild(frag, node);
    });
  }

  function clearHighlights(root) {
    root.querySelectorAll('mark.sh').forEach(m => {
      
      const child = m.firstChild;
      if (child?.classList?.contains('vt-hash')) {
        m.parentNode.insertBefore(child, m);
        m.remove();
      } else {
        m.replaceWith(document.createTextNode(m.textContent));
      }
    });
    root.querySelectorAll('.kv-val.hash, .kv-val').forEach(n => n.normalize());
  }

  function apply() {

    const counts = { alert:0, warning:0, notice:0, total:0, visible:0 };
    const modCounts = new Map(); 

    $$('.finding').forEach(f => {
      clearHighlights(f);
      const c = _cache.get(f);
      if (!c) return;
      counts.total++;

      const matchBase =
        (!S || c.txt.includes(S)) &&
        (!inc.length || inc.some(k => c.path.includes(k.trim().toLowerCase()) || c.txt.includes(k.trim().toLowerCase()))) &&
        (!exc.length || !exc.some(k => c.path.includes(k.trim().toLowerCase()) || c.txt.includes(k.trim().toLowerCase())));

      if (!matchBase) { f.style.display='none'; return; }

      const matchTab = F==='all' || c.sev===F || (F==='warning' && (c.sev==='warn'||c.sev==='warning'));
      const sev = (c.sev==='warn'||c.sev==='warning') ? 'warning' : (c.sev==='high' ? 'alert' : c.sev);
      if (sev in counts) counts[sev]++;

      if (matchTab) modCounts.set(c.mod, (modCounts.get(c.mod)||0) + 1);

      const matchMod = !M || c.mod === M.toLowerCase();
      const vis = matchTab && matchMod;
      f.style.display = vis ? '' : 'none';

      if (vis) {
        counts.visible++;
        if (S) {
          highlightIn(f, S);
          if (f.querySelector('mark.sh')) f.classList.add('expanded');
        }
      }
    });

    el('cnt-alert')  && (el('cnt-alert').textContent  = counts.alert);
    el('cnt-warn')   && (el('cnt-warn').textContent   = counts.warning);
    el('cnt-notice') && (el('cnt-notice').textContent = counts.notice);
    el('cnt-all')    && (el('cnt-all').textContent    = counts.total);

    $$('.mod-chip[data-mod]').forEach(chip => {
      const cnt = chip.querySelector('.mod-count');
      if (cnt) cnt.textContent = modCounts.get(chip.dataset.mod) || 0;
    });

    const sc = el('showing-count');
    if (sc) {
      sc.innerHTML = counts.visible === counts.total
        ? '<span>' + counts.total + '</span> findings'
        : 'Showing <span>' + counts.visible + '</span> of <span>' + counts.total + '</span>';
    }

    const findingsEl = el('findings');
    let empty = findingsEl.querySelector('.empty-state');
    if (counts.visible === 0) {
      if (!empty) {
        empty = document.createElement('div');
        empty.className = 'empty-state';
        empty.textContent = 'No findings match the current filters';
        findingsEl.appendChild(empty);
      }
    } else {
      empty?.remove();
    }

    renderPanel('active-filters-panel', inc, 'filter-chip',   'FILTER IN', 'removeInc', 'clearInc', 'kw-input', 'handleInc');
    renderPanel('exclude-filters-panel', exc, 'exclude-chip', 'FILTER OUT', 'removeExc', 'clearExc', 'ex-input', 'handleExc');
  }

  function renderPanel(panelId, arr, chipCls, label, removeFn, clearFn, inputId, kwFn) {
    const p = el(panelId); if(!p) return;
    const isExc = chipCls==='exclude-chip';
    const dotColor = isExc ? 'var(--al)' : arr.length ? 'var(--ac)' : 'var(--tx2)';
    const chips = arr.map((kw,i)=>{
      const lbl = kw.length>28 ? kw.slice(0,26)+'…' : kw;
      const prefix = isExc ? '<span class="ex-prefix">✗</span>' : '';
      return '<span class="' + chipCls + '">' + prefix + lbl + '<span class="x" onclick="' + removeFn + '(' + i + ')">✕</span></span>';
    }).join('');
    p.innerHTML =
      '<span class="af-label"><span class="af-dot" style="background:' + dotColor + '"></span>' + label + '</span>' +
      chips +
      '<span class="' + chipCls + ' add-chip"><input id="' + inputId + '" class="kw-input" type="text"' +
      ' placeholder="+ add ' + (isExc ? 'exclude' : 'filter') + '\u2026" onkeydown="' + kwFn + '(event)"/></span>' +
      '<div class="af-actions"><button class="af-btn" onclick="' + clearFn + '()">Clear All</button></div>';
  }

  function kwHandler(e, arr) {
    if(e.key==='Enter'){
      const v=e.target.value.trim();
      if(v&&!arr.includes(v)){
        arr.push(v);
        if(arr===inc) S=v.toLowerCase();
        apply();
      } else e.target.value='';
    }
    if(e.key==='Escape') e.target.value='';
  }
  function handleInc(e){kwHandler(e,inc)}
  function handleExc(e){kwHandler(e,exc)}
  function removeInc(i){
    const removed = inc[i];
    inc.splice(i,1);
    if (S === removed) S = '';
    apply();
  }
  function removeExc(i){exc.splice(i,1);apply()}
  function clearInc(){inc=[];S='';apply()}
  function clearExc(){exc=[];apply()}

  function filterFindings(type,btn){
    $$('.tab').forEach(t=>t.classList.remove('active'));
    btn.classList.add('active'); F=type; apply();
  }
  function filterModule(mod,chip){
    M=(M===mod)?'':mod;
    $$('.mod-chip').forEach(c=>c.classList.remove('active-mod'));
    if(M) chip.classList.add('active-mod');
    apply();
  }

  function expandAll(){$$('.finding').forEach(f=>{if(f.style.display!=='none')f.classList.add('expanded')})}
  function collapseAll(){$$('.finding').forEach(f=>f.classList.remove('expanded'));}
  function toggleFinding(h){h.closest('.finding').classList.toggle('expanded')}

  function sortFindings() {
    const container = el('findings');
    [...$$('.finding')].sort((a,b) => (_cache.get(b)?.score||0) - (_cache.get(a)?.score||0))
      .forEach(n => container.appendChild(n));
  }

  document.addEventListener('DOMContentLoaded', () => {
    
    const VT = 'https://www.virustotal.com/gui/file/';
    $$('.kv-val.hash').forEach(span => {
      const key = span.previousElementSibling?.textContent?.trim() || '';
      if (/^(MD5|SHA1|SHA256|Image MD5|Image SHA1|Image SHA256)$/i.test(key)) {
        const hash = span.textContent.trim();
        const a = document.createElement('a');
        a.href = VT + hash; a.target = '_blank'; a.rel = 'noopener noreferrer';
        a.className = 'vt-hash'; a.textContent = hash;
        span.textContent = ''; span.appendChild(a);
      }
    });

    buildCache();
    sortFindings();
    apply();

    const tip = document.createElement('div');
    tip.id = 'sel-tip';
    tip.innerHTML = '<button class="tip-in" onclick="selFilterIn()">＋ Filter In</button><div class="tip-div"></div><button class="tip-out" onclick="selFilterOut()">－ Filter Out</button>';
    document.body.appendChild(tip);

    let _selText = '';

    function hideTip() {
      tip.style.display = 'none';
      _selText = '';
    }

    document.addEventListener('mouseup', e => {
      if (e.target.closest('#sel-tip')) return;
      const sel = window.getSelection();
      if (!sel || !sel.toString().trim()) { hideTip(); return; }
      const anchor = sel.anchorNode?.parentElement?.closest('.finding-body');
      if (!anchor) { hideTip(); return; }

      const txt = sel.toString().trim();
      if (!txt || txt.length < 2) { hideTip(); return; }
      _selText = txt;

      const range = sel.getRangeAt(0);
      const endRange = range.cloneRange();
      endRange.collapse(false); 
      const rect = endRange.getBoundingClientRect();

      tip.style.visibility = 'hidden';
      tip.style.display = 'flex';
      const tipW = tip.offsetWidth;
      const tipH = tip.offsetHeight;
      tip.style.visibility = '';

      let x = rect.left;
      let y = rect.bottom + 6;
      if (rect.bottom + tipH + 8 > window.innerHeight) y = rect.top - tipH - 6;
      x = Math.max(8, Math.min(x, window.innerWidth - tipW - 8));

      tip.style.left = x + 'px';
      tip.style.top  = y + 'px';
    });

    document.addEventListener('selectionchange', () => {
      const sel = window.getSelection();
      if (!sel || sel.isCollapsed) hideTip();
    });

    window.selFilterIn = () => {
      const v = _selText.trim().toLowerCase();
      if (v && !inc.includes(v)) {
        inc.push(v);
        S = v;
        apply();
      }
      hideTip(); window.getSelection()?.removeAllRanges();
    };
    window.selFilterOut = () => {
      const v = _selText.trim().toLowerCase();
      if (v && !exc.includes(v)) { exc.push(v); apply(); }
      hideTip(); window.getSelection()?.removeAllRanges();
    };

  });</script>
</body>
</html>
`
}
