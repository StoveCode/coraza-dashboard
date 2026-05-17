package tailer

import (
	"context"
	"time"

	"github.com/nxadm/tail"
	"github.com/rs/zerolog/log"

	"github.com/corazawaf/coraza-dashboard/internal/db"
	"github.com/corazawaf/coraza-dashboard/internal/parser"
	"github.com/jackc/pgx/v5/pgxpool"
)

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
		Poll:      true, // use polling instead of inotify to avoid container crashes
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
	event := parser.ParseLine(raw)
	if event == nil {
		return
	}

	if err := db.InsertEvent(ctx, t.pool, event); err != nil {
		log.Error().Err(err).Msg("failed to insert WAF event")
	} else {
		log.Debug().Str("unique_id", event.UniqueID).Int("rule_id", event.RuleID).Bool("disruptive", event.Disruptive).Msg("inserted WAF event")
	}
}
