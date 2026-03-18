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
	tmpl := template.Must(template.New("report").Parse(Template()))
	return tmpl, nil
}
