package configwriter

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"
	"text/template"

	"github.com/corazawaf/coraza-dashboard/internal/models"
)

const yamlTemplate = `bind: 0.0.0.0:9000
log_level: info
log_file: /dev/stdout
log_format: json

default_application: sample_app

applications:
  - name: sample_app
    directives: |
      Include @coraza.conf-recommended
      Include @crs-setup.conf.example
{{ if gt .ParanoiaLevel 0 }}
      # Paranoia Level
      SecAction "id:900000,phase:1,nolog,pass,t:none,setvar:tx.blocking_paranoia_level={{ .ParanoiaLevel }}"
{{ end }}{{ if .CustomThresholds }}
      # Anomaly Score Thresholds
      SecAction "id:900110,phase:1,nolog,pass,t:none,setvar:tx.inbound_anomaly_score_threshold={{ .InboundThreshold }},setvar:tx.outbound_anomaly_score_threshold={{ .OutboundThreshold }}"
{{ end }}
      Include @owasp_crs/*.conf
      SecRuleEngine {{ .EngineMode }}
{{ if .DisabledRuleIds }}
      # Disabled Rules
      SecRuleRemoveById {{ .DisabledRuleIdsStr }}
{{ end }}{{ if .DisabledTags }}
      # Disabled Tags
{{ range .DisabledTagsList }}      SecRuleRemoveByTag "{{ . }}"
{{ end }}{{ end }}
    response_check: {{ if .ResponseCheck }}true{{ else }}false{{ end }}
    transaction_ttl_ms: 60000
    log_level: info
    log_file: /dev/stdout
    log_format: json
`

type templateData struct {
	EngineMode           string
	ParanoiaLevel        int
	ParanoiaLevelEnabled bool
	InboundThreshold     int
	OutboundThreshold    int
	CustomThresholds     bool
	DisabledRuleIds      bool
	DisabledRuleIdsStr   string
	DisabledTags         bool
	DisabledTagsList     []string
	ResponseCheck        bool
}

func WriteConfig(cfg *models.RulesConfig) error {
	configPath := os.Getenv("CORAZA_CONFIG_PATH")
	if configPath == "" {
		return fmt.Errorf("CORAZA_CONFIG_PATH env var not set")
	}

	content, err := GenerateYAML(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, []byte(content), 0644)
}

var ruleIDSanitizer = regexp.MustCompile(`^\d+$`)

func sanitizeRuleIDs(ids []string) []string {
	result := make([]string, 0, len(ids))
	for _, id := range ids {
		if ruleIDSanitizer.MatchString(id) {
			result = append(result, id)
		}
	}
	return result
}

func GenerateYAML(cfg *models.RulesConfig) (string, error) {
	sanitizedIDs := sanitizeRuleIDs(cfg.DisabledRuleIds)
	data := templateData{
		EngineMode:           cfg.EngineMode,
		ParanoiaLevel:        cfg.ParanoiaLevel,
		ParanoiaLevelEnabled: cfg.ParanoiaLevelEnabled,
		InboundThreshold:     cfg.InboundThreshold,
		OutboundThreshold: cfg.OutboundThreshold,
		CustomThresholds:  cfg.InboundThreshold != 5 || cfg.OutboundThreshold != 4,
		DisabledRuleIds:   len(sanitizedIDs) > 0,
		DisabledRuleIdsStr: strings.Join(sanitizedIDs, " "),
		DisabledTags:      len(cfg.DisabledTags) > 0,
		DisabledTagsList:  cfg.DisabledTags,
		ResponseCheck:     cfg.ResponseCheck,
	}

	tmpl, err := template.New("coraza").Parse(yamlTemplate)
	if err != nil {
		return "", fmt.Errorf("template parse error: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("template execute error: %w", err)
	}

	return buf.String(), nil
}
