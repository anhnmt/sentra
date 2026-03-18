package report

// Header returns the HTML header section.
func Header() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Sentra Scan Report</title>
<style>
* { margin: 0; padding: 0; box-sizing: border-box; }
body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; background: #f5f7fa; color: #333; line-height: 1.6; }
.container { max-width: 1400px; margin: 0 auto; padding: 20px; }
.header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 30px; border-radius: 12px; margin-bottom: 24px; }
.header h1 { font-size: 28px; margin-bottom: 8px; }
.header .subtitle { opacity: 0.9; font-size: 14px; }
.navbar { background: white; border-radius: 12px; padding: 16px 24px; margin-bottom: 24px; display: flex; gap: 24px; box-shadow: 0 2px 8px rgba(0,0,0,0.06); }
.navbar a { color: #667eea; text-decoration: none; font-weight: 500; padding: 8px 16px; border-radius: 8px; transition: all 0.2s; }
.navbar a:hover { background: #f0f4ff; }
.summary-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(240px, 1fr)); gap: 20px; margin-bottom: 24px; }
.card { background: white; border-radius: 12px; padding: 24px; box-shadow: 0 2px 8px rgba(0,0,0,0.06); }
.card-label { font-size: 13px; color: #888; text-transform: uppercase; letter-spacing: 0.5px; margin-bottom: 8px; }
.card-value { font-size: 32px; font-weight: 700; color: #1a1a2e; }
.card-value.alert { color: #e74c3c; }
.card-value.warning { color: #f39c12; }
.card-value.notice { color: #3498db; }
.card-value.success { color: #27ae60; }
.cmd-value { font-family: "SF Mono", Monaco, monospace; font-size: 13px; background: #f8f9fa; padding: 12px; border-radius: 8px; border: 1px solid #eee; word-break: break-all; }
.section { background: white; border-radius: 12px; padding: 24px; margin-bottom: 24px; box-shadow: 0 2px 8px rgba(0,0,0,0.06); }
.section h2 { font-size: 20px; margin-bottom: 20px; padding-bottom: 12px; border-bottom: 2px solid #f0f4ff; }
.table-wrapper { overflow-x: auto; }
table { width: 100%; border-collapse: collapse; }
th, td { padding: 14px 16px; text-align: left; border-bottom: 1px solid #eee; }
th { background: #f8f9fa; font-weight: 600; color: #555; font-size: 13px; text-transform: uppercase; }
tr:hover { background: #f8f9fa; }
.badge { display: inline-block; padding: 4px 10px; border-radius: 20px; font-size: 12px; font-weight: 600; }
.badge.alert { background: #fee; color: #e74c3c; }
.badge.warning { background: #fff3e0; color: #f39c12; }
.badge.notice { background: #e3f2fd; color: #3498db; }
.footer { text-align: center; padding: 24px; color: #888; font-size: 13px; }
.severity-icon { width: 12px; height: 12px; border-radius: 50%; display: inline-block; margin-right: 8px; }
.severity-icon.alert { background: #e74c3c; }
.severity-icon.warning { background: #f39c12; }
.severity-icon.notice { background: #3498db; }
.string-item { display: inline-block; background: #f0f4ff; padding: 2px 8px; border-radius: 4px; margin: 2px; font-family: monospace; font-size: 12px; }
</style>
</head>
<body>`
}

// Navbar returns the navigation bar section.
func Navbar() string {
	return `<div class="navbar">
<a href="#summary">Summary</a>
<a href="#findings">Findings</a>
<a href="#system">System Info</a>
</div>`
}

// SummarySection returns the summary cards section.
func SummarySection() string {
	return `{{define "summary"}}
<div class="section" id="summary">
<h2>Scan Summary</h2>
<div class="summary-grid">
<div class="card">
<div class="card-label">Total Scanned</div>
<div class="card-value">{{.Scanned}}</div>
</div>
<div class="card">
<div class="card-label">Matches Found</div>
<div class="card-value alert">{{.MatchCount}}</div>
</div>
<div class="card">
<div class="card-label">Errors</div>
<div class="card-value warning">{{.ErrorCount}}</div>
</div>
<div class="card">
<div class="card-label">Skipped</div>
<div class="card-value">{{.Skipped}}</div>
</div>
<div class="card">
<div class="card-label">Duration</div>
<div class="card-value">{{.Duration}}</div>
</div>
<div class="card">
<div class="card-label">Status</div>
<div class="card-value success">{{.Status}}</div>
</div>
</div>
</div>
{{end}}`
}

// SystemInfoSection returns the system info section.
func SystemInfoSection() string {
	return `{{define "system"}}
<div class="section" id="system">
<h2>System Information</h2>
<div class="summary-grid">
<div class="card">
<div class="card-label">Hostname</div>
<div class="card-value" style="font-size:20px">{{.Hostname}}</div>
</div>
<div class="card">
<div class="card-label">Operating System</div>
<div class="card-value" style="font-size:20px">{{.OS}} {{.Arch}}</div>
</div>
<div class="card">
<div class="card-label">IP Address</div>
<div class="card-value" style="font-size:20px">{{.IPAddr}}</div>
</div>
<div class="card">
<div class="card-label">User</div>
<div class="card-value" style="font-size:20px">{{.User}}</div>
</div>
</div>
<div class="card" style="margin-top:20px">
<div class="card-label">Command Line</div>
<div class="cmd-value">{{.CommandLine}}</div>
</div>
</div>
{{end}}`
}

// FindingsSection returns the findings table section.
func FindingsSection() string {
	return `{{define "findings"}}
<div class="section" id="findings">
<h2>Detection Findings</h2>
<div class="summary-grid">
<div class="card">
<div class="card-label">Alert</div>
<div class="card-value alert">{{.AlertCount}}</div>
</div>
<div class="card">
<div class="card-label">Warning</div>
<div class="card-value warning">{{.WarningCount}}</div>
</div>
<div class="card">
<div class="card-label">Notice</div>
<div class="card-value notice">{{.NoticeCount}}</div>
</div>
</div>
<div class="table-wrapper">
<table>
<thead>
<tr>
<th>Severity</th>
<th>Rule</th>
<th>Target</th>
<th>Module</th>
<th>Description</th>
</tr>
</thead>
<tbody>
{{range .Findings}}
<tr>
<td><span class="badge {{.Severity}}">{{.Severity}}</span></td>
<td><strong>{{.RuleName}}</strong></td>
<td>{{.Target}}</td>
<td>{{.Module}}</td>
<td>{{.Description}}</td>
</tr>
{{end}}
</tbody>
</table>
</div>
</div>
{{end}}`
}

// Footer returns the HTML footer section.
func Footer() string {
	return `<div class="footer">
<p>Generated by Sentra v{{.Version}} | Scan ID: {{.ScanID}}</p>
<p>Report generated at {{.GeneratedAt}}</p>
</div>
</body>
</html>`
}
