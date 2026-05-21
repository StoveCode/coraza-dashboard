package catalog

import (
	"io/fs"
	"regexp"
	"strconv"
	"strings"

	crs "github.com/corazawaf/coraza-coreruleset/v4"
)

// GetCRSVersion extracts the CRS version from the embedded crs-setup.conf.example.
// Returns a version string like "v4.0.0-rc2" or "unknown" as fallback.
func GetCRSVersion() string {
	data, err := fs.ReadFile(crs.FS, "@crs-setup.conf.example")
	if err != nil {
		return "unknown"
	}
	// Match version patterns like "ver.4.0.0-rc2" or "v4.0.0"
	re := regexp.MustCompile(`(?i)(?:ver\.|v)(\d+\.\d+\.\d+[\w.-]*)`)
	m := re.FindStringSubmatch(string(data))
	if m != nil {
		return "v" + m[1]
	}
	return "unknown"
}

// Rule represents a single OWASP CRS rule with its key metadata.
type Rule struct {
	ID            int    `json:"id"`
	Msg           string `json:"msg"`
	ParanoiaLevel int    `json:"paranoia_level"`
	Tag           string `json:"tag"`
	Severity      string `json:"severity"`
	Directive     string `json:"directive"`
}

var (
	chunkSplitRe = regexp.MustCompile(`(?m)^(?:SecRule|SecAction)\b`)
	ruleIDRe     = regexp.MustCompile(`\bid:(\d{6})\b`)
	msgRe        = regexp.MustCompile(`msg:'([^']+)'`)
	attackTagRe  = regexp.MustCompile(`tag:'(attack-[^']+)'`)
	plRe         = regexp.MustCompile(`tag:'paranoia-level/(\d)'`)
	sevRe        = regexp.MustCompile(`(?i)severity:'(CRITICAL|ERROR|WARNING|NOTICE)'`)
)

// Load parses all embedded CRS .conf files and returns rules that have an attack- tag.
func Load() ([]Rule, error) {
	var rules []Rule
	seen := map[int]bool{}

	err := fs.WalkDir(crs.FS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".conf") {
			return err
		}
		data, readErr := fs.ReadFile(crs.FS, path)
		if readErr != nil {
			return nil // skip unreadable files
		}
		for _, r := range parseConf(string(data)) {
			if !seen[r.ID] {
				seen[r.ID] = true
				rules = append(rules, r)
			}
		}
		return nil
	})
	return rules, err
}

func parseConf(content string) []Rule {
	locs := chunkSplitRe.FindAllStringIndex(content, -1)
	if len(locs) == 0 {
		return nil
	}

	var rules []Rule
	for i, loc := range locs {
		end := len(content)
		if i+1 < len(locs) {
			end = locs[i+1][0]
		}
		chunk := content[loc[0]:end]

		idM := ruleIDRe.FindStringSubmatch(chunk)
		if idM == nil {
			continue
		}
		id, _ := strconv.Atoi(idM[1])

		tagM := attackTagRe.FindStringSubmatch(chunk)
		if tagM == nil {
			continue
		}

		msg := ""
		if m := msgRe.FindStringSubmatch(chunk); m != nil {
			msg = m[1]
		}
		pl := 1
		if m := plRe.FindStringSubmatch(chunk); m != nil {
			pl, _ = strconv.Atoi(m[1])
		}
		sev := "notice"
		if m := sevRe.FindStringSubmatch(chunk); m != nil {
			sev = strings.ToLower(m[1])
		}

		rules = append(rules, Rule{
			ID:            id,
			Msg:           msg,
			ParanoiaLevel: pl,
			Tag:           tagM[1],
			Severity:      sev,
			Directive:     cleanDirective(chunk),
		})
	}
	return rules
}

func cleanDirective(s string) string {
	// Resolve backslash line continuations
	s = regexp.MustCompile(`\\\s*\n\s*`).ReplaceAllString(s, " ")
	// Collapse multiple whitespace into single space
	s = regexp.MustCompile(`\s{2,}`).ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

// TagLabel returns a human-readable label for a CRS attack tag.
func TagLabel(tag string) string {
	m := map[string]string{
		"attack-sqli":               "SQL Injection",
		"attack-xss":                "Cross-Site Scripting",
		"attack-rce":                "Remote Code Execution",
		"attack-lfi":                "Local File Inclusion",
		"attack-rfi":                "Remote File Inclusion",
		"attack-ssrf":               "Server-Side Request Forgery",
		"attack-ssti":               "Template Injection",
		"attack-injection":          "Generic Injection",
		"attack-injection-generic":  "Generic Injection",
		"attack-injection-java":     "Java Injection",
		"attack-injection-php":      "PHP Injection",
		"attack-scanner":            "Scanner Detection",
		"attack-reputation-scanner": "Scanner Reputation",
		"attack-protocol":           "Protocol Attacks",
		"attack-multipart-header":   "Multipart Header Attacks",
		"attack-fixation":           "Session Fixation",
		"attack-disclosure":         "Information Disclosure",
		"attack-generic":            "Generic Enforcement",
		"attack-deprecated-header":  "Deprecated Headers",
	}
	if l, ok := m[tag]; ok {
		return l
	}
	return toTitle(strings.ReplaceAll(strings.TrimPrefix(tag, "attack-"), "-", " "))
}

func toTitle(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// TagDescription returns a description for a CRS attack tag.
func TagDescription(tag string) string {
	m := map[string]string{
		"attack-sqli":               "SQL injection via libinjection and pattern matching",
		"attack-xss":                "Cross-Site Scripting via libinjection and pattern matching",
		"attack-rce":                "OS command injection, RCE, web shells",
		"attack-lfi":                "Path traversal and local file inclusion",
		"attack-rfi":                "Remote file inclusion attempts",
		"attack-ssrf":               "SSRF including cloud metadata access",
		"attack-ssti":               "Server-side template injection",
		"attack-injection":          "Language-agnostic injection attacks",
		"attack-injection-generic":  "Language-agnostic injection attacks",
		"attack-injection-java":     "Java-specific injection attacks",
		"attack-injection-php":      "PHP-specific injection attacks",
		"attack-scanner":            "Web application scanner user-agents",
		"attack-reputation-scanner": "Known malicious scanner signatures",
		"attack-protocol":           "HTTP protocol violations and smuggling",
		"attack-multipart-header":   "Malformed multipart/form-data headers",
		"attack-fixation":           "Session fixation and hijacking",
		"attack-disclosure":         "Sensitive data leakage in responses",
		"attack-generic":            "Method enforcement and common exceptions",
		"attack-deprecated-header":  "Deprecated or dangerous HTTP headers",
	}
	if d, ok := m[tag]; ok {
		return d
	}
	return ""
}
