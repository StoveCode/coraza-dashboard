package tailer

import (
	"context"
	"encoding/json"
	"time"

	"github.com/nxadm/tail"
	"github.com/rs/zerolog/log"

	"github.com/corazawaf/coraza-dashboard/internal/db"
	"github.com/corazawaf/coraza-dashboard/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

// logLine is the top-level structure of each coraza-spoa JSON log entry.
type logLine struct {
	Level string       `json:"level"`
	Match *matchedRule `json:"match"`
	Time  time.Time    `json:"time"`
}

// matchedRule maps to the "match" field in coraza-spoa log output.
type matchedRule struct {
	Client     string   `json:"client"`
	File       string   `json:"file"`
	Line       int      `json:"line"`
	RuleID     int      `json:"rule_id"`
	Msg        string   `json:"msg"`
	Data       string   `json:"data"`
	Severity   string   `json:"severity"`
	SeverityID int      `json:"severity_id"`
	Tags       []string `json:"tags"`
	Server     string   `json:"server"`
	URI        string   `json:"uri"`
	UniqueID   string   `json:"unique_id"`
	Disruptive bool     `json:"disruptive"`
	PhaseID    int      `json:"phase_id"`
	Phase      string   `json:"phase"`
}

// Tailer follows a log file and inserts WAF events into the database.
type Tailer struct {
	logFile string
	pool    *pgxpool.Pool
}

func New(logFile string, pool *pgxpool.Pool) *Tailer {
	return &Tailer{logFile: logFile, pool: pool}
}

func (t *Tailer) Run(ctx context.Context) {
	cfg := tail.Config{
		Follow:    true,
		ReOpen:    true,
		MustExist: false,
		Logger:    tail.DiscardingLogger,
	}

	log.Info().Str("file", t.logFile).Msg("starting log tailer")

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		tl, err := tail.TailFile(t.logFile, cfg)
		if err != nil {
			log.Warn().Err(err).Str("file", t.logFile).Msg("tail error, retrying in 5s")
			time.Sleep(5 * time.Second)
			continue
		}

		for line := range tl.Lines {
			if line.Err != nil {
				log.Warn().Err(line.Err).Msg("tail line error")
				continue
			}
			t.processLine(ctx, []byte(line.Text))
		}

		// Channel closed (file gone?), retry
		log.Warn().Str("file", t.logFile).Msg("tail channel closed, reopening")
		time.Sleep(2 * time.Second)
	}
}

func (t *Tailer) processLine(ctx context.Context, raw []byte) {
	var ll logLine
	if err := json.Unmarshal(raw, &ll); err != nil {
		log.Debug().Err(err).Msg("failed to parse log line")
		return
	}

	if ll.Match == nil {
		// Normal info/debug log, not a rule match
		return
	}

	m := ll.Match
	event := &models.WAFEvent{
		Timestamp:  ll.Time,
		Client:     m.Client,
		Server:     m.Server,
		URI:        m.URI,
		RuleID:     m.RuleID,
		RuleMsg:    m.Msg,
		RuleFile:   m.File,
		Severity:   m.Severity,
		SeverityID: m.SeverityID,
		Phase:      m.Phase,
		PhaseID:    m.PhaseID,
		Disruptive: m.Disruptive,
		Tags:       m.Tags,
		Data:       m.Data,
		UniqueID:   m.UniqueID,
		RawLog:     raw,
	}

	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}

	if err := db.InsertEvent(ctx, t.pool, event); err != nil {
		log.Error().Err(err).Msg("failed to insert WAF event")
	} else {
		log.Debug().Str("unique_id", m.UniqueID).Int("rule_id", m.RuleID).Bool("disruptive", m.Disruptive).Msg("inserted WAF event")
	}
}
