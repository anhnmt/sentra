package report

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/anhnmt/sentra/internal/store"
)

const defaultVersion = "0.1.0"

// Generator creates HTML reports from bbolt database.
type Generator struct {
	store   *store.Store
	dbPath  string
	version string
	ipAddr  string
	user    string
}

// Option configures a Generator.
type Option func(*Generator)

// WithVersion sets the report version.
func WithVersion(v string) Option {
	return func(g *Generator) { g.version = v }
}

// WithIPAddr sets the IP address for the report.
func WithIPAddr(ip string) Option {
	return func(g *Generator) { g.ipAddr = ip }
}

// WithUser sets the user for the report.
func WithUser(u string) Option {
	return func(g *Generator) { g.user = u }
}

// NewGenerator creates a new report generator.
func NewGenerator(dbPath string, opts ...Option) *Generator {
	g := &Generator{
		dbPath:  dbPath,
		version: defaultVersion,
	}
	for _, opt := range opts {
		opt(g)
	}
	return g
}

// NewGeneratorWithStore creates a new report generator with an existing store.
func NewGeneratorWithStore(s *store.Store, opts ...Option) *Generator {
	g := &Generator{
		store:   s,
		version: defaultVersion,
	}
	for _, opt := range opts {
		opt(g)
	}
	return g
}

// Generate creates an HTML report for the given scan ID.
func (g *Generator) Generate(scanID, outputPath, commandLine string) error {
	// Use existing store if provided, otherwise open new connection
	var s *store.Store
	var err error
	if g.store != nil {
		s = g.store
	} else {
		s, err = store.Open(g.dbPath)
		if err != nil {
			return fmt.Errorf("open db: %w", err)
		}
		defer s.Close()
	}

	// Load scan record
	scan, err := s.GetScan(scanID)
	if err != nil {
		return fmt.Errorf("get scan: %w", err)
	}
	if scan == nil {
		return fmt.Errorf("scan not found: %s", scanID)
	}

	// Load device info
	device, _ := s.GetDevice(scan.Target)
	if device == nil {
		device = &store.DeviceInfo{Hostname: "unknown", OS: "unknown"}
	}

	// Load matches
	matches, err := s.GetMatches(scanID)
	if err != nil {
		return fmt.Errorf("get matches: %w", err)
	}

	// Build report data
	data := ReportData{
		Version:     g.version,
		ScanID:      scanID,
		GeneratedAt: time.Now().UTC(),
		Hostname:    device.Hostname,
		OS:          device.OS,
		Arch:        device.Arch,
		IPAddr:      g.ipAddr,
		User:        g.user,
		ScanStart:   scan.StartedAt,
		ScanEnd:     scan.FinishedAt,
		Duration:    scan.FinishedAt.Sub(scan.StartedAt),
		Target:      scan.Target,
		RulesDir:    scan.RulesDir,
		Scanned:     scan.Scanned,
		Skipped:     scan.Skipped,
		MatchCount:  scan.MatchCount,
		ErrorCount:  scan.ErrorCount,
		Status:      scan.Status,
		Workers:     scan.Workers,
		CommandLine: commandLine,
		Findings:    g.buildFindings(matches),
	}

	// Count by severity
	for _, f := range data.Findings {
		switch f.Severity {
		case "alert":
			data.AlertCount++
		case "warning":
			data.WarningCount++
		case "notice":
			data.NoticeCount++
		}
	}

	// Render HTML using embedded template
	tmpl, err := LoadTemplate()
	if err != nil {
		return fmt.Errorf("load template: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("execute template: %w", err)
	}

	// Write to file
	if err := os.WriteFile(outputPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	fmt.Printf("Report written to: %s\n", outputPath)
	return nil
}

// buildFindings converts match records to report findings.
func (g *Generator) buildFindings(matches []store.MatchRecord) []Finding {
	findings := make([]Finding, 0, len(matches))

	for i, m := range matches {
		f := Finding{
			ID:       fmt.Sprintf("finding-%d", i+1),
			Severity: g.inferSeverity(m),
			Score:    g.inferScore(m),
			Module:   m.DetectorName,
			Target:   m.Target,
			RuleName: m.RuleName,
			RuleType: "YARA Rule",
			FileType: g.inferFileType(m.Target),
		}

		// Extract metadata
		if m.Metadata != nil {
			if v, ok := m.Metadata["description"].(string); ok {
				f.Description = v
			}
			if v, ok := m.Metadata["author"].(string); ok {
				f.Author = v
			}
			if v, ok := m.Metadata["date"].(string); ok {
				f.Date = v
			}
			if v, ok := m.Metadata["class"].(string); ok {
				f.Class = v
			}
			if v, ok := m.Metadata["md5"].(string); ok {
				f.MD5 = v
			}
			if v, ok := m.Metadata["sha256"].(string); ok {
				f.SHA256 = v
			}
			if v, ok := m.Metadata["sha1"].(string); ok {
				f.SHA1 = v
			}
			if v, ok := m.Metadata["strings"].([]any); ok {
				f.Strings = g.extractStrings(v)
			}
			if v, ok := m.Metadata["references"].([]any); ok {
				f.Refs = g.extractStringsToStr(v)
			}
			if v, ok := m.Metadata["tags"].([]any); ok {
				f.AttackTags = g.extractAttackTags(v)
			}
		}

		findings = append(findings, f)
	}

	return findings
}

// inferSeverity determines severity from match metadata.
func (g *Generator) inferSeverity(m store.MatchRecord) string {
	if m.Metadata == nil {
		return "warning"
	}
	if v, ok := m.Metadata["severity"].(string); ok {
		return v
	}
	if v, ok := m.Metadata["score"].(float64); ok {
		if v >= 80 {
			return "alert"
		}
		if v >= 60 {
			return "warning"
		}
		return "notice"
	}
	return "warning"
}

// inferScore extracts score from metadata.
func (g *Generator) inferScore(m store.MatchRecord) int {
	if m.Metadata == nil {
		return 50
	}
	if v, ok := m.Metadata["score"].(float64); ok {
		return int(v)
	}
	if v, ok := m.Metadata["subscore"].(float64); ok {
		return int(v)
	}
	return 50
}

// inferFileType determines the file type from target path.
func (g *Generator) inferFileType(target string) string {
	// Simple heuristics based on extension
	switch {
	case contains(target, ".exe"):
		return "PE EXECUTABLE"
	case contains(target, ".dll"):
		return "DLL"
	case contains(target, ".elf"):
		return "ELF EXECUTABLE"
	case contains(target, ".zip"), contains(target, ".tar"), contains(target, ".gz"):
		return "ARCHIVE"
	case contains(target, ".ps1"), contains(target, ".bat"), contains(target, ".cmd"):
		return "SCRIPT"
	case contains(target, "HKLM"), contains(target, "HKCU"):
		return "REGISTRY KEY"
	case contains(target, "EventID"):
		return "EVENT LOG"
	default:
		return "FILE"
	}
}

// extractStrings converts metadata strings to MatchedString.
func (g *Generator) extractStrings(v []any) []MatchedString {
	result := make([]MatchedString, 0, len(v))
	for _, s := range v {
		if str, ok := s.(string); ok {
			result = append(result, MatchedString{Content: str})
		}
	}
	return result
}

// extractStringsToStr converts metadata strings to []string.
func (g *Generator) extractStringsToStr(v []any) []string {
	result := make([]string, 0, len(v))
	for _, s := range v {
		if str, ok := s.(string); ok {
			result = append(result, str)
		}
	}
	return result
}

// extractAttackTags extracts MITRE ATT&CK tags.
func (g *Generator) extractAttackTags(v []any) []string {
	result := make([]string, 0, len(v))
	for _, s := range v {
		if str, ok := s.(string); ok {
			if contains(str, "attack.") || contains(str, "T") {
				result = append(result, str)
			}
		}
	}
	return result
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsAt(s, substr))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// GetScanList returns all scan IDs from the database.
func (g *Generator) GetScanList() ([]string, error) {
	s, err := store.Open(g.dbPath)
	if err != nil {
		return nil, err
	}
	defer s.Close()
	return s.GetAllScanIDs()
}

// GenerateLatest generates a report for the most recent scan.
func (g *Generator) GenerateLatest(outputPath, commandLine string) error {
	ids, err := g.GetScanList()
	if err != nil {
		return err
	}
	if len(ids) == 0 {
		return fmt.Errorf("no scans found")
	}
	// Latest scan is the last one (most recent timestamp)
	return g.Generate(ids[len(ids)-1], outputPath, commandLine)
}
