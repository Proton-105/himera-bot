package database

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/himera-bot/trading-bot/pkg/logger"
)

// Migrator applies plain .sql file migrations in lexical order.
// Phase0: only .up.sql is supported.
type Migrator struct {
	db  *sql.DB
	log *logger.Logger
}

func NewMigrator(db *sql.DB, log *logger.Logger) *Migrator {
	return &Migrator{
		db:  db,
		log: log,
	}
}

func (m *Migrator) baseLogger() *logger.Logger {
	if m.log != nil {
		return m.log
	}

	l := logger.New()
	m.log = l
	return l
}

// ApplyDir scans dir, finds *.up.sql, sorts them, and executes sequentially.
func (m *Migrator) ApplyDir(ctx context.Context, dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read migrations dir %q: %w", dir, err)
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if isUpMigration(name) {
			files = append(files, filepath.Join(dir, name))
		}
	}

	baseLog := m.baseLogger().With(slog.String("dir", dir))

	if len(files) == 0 {
		baseLog.Info("no .up.sql migrations found")
		return nil
	}

	sort.Strings(files)

	for _, path := range files {
		if err := m.applyFile(ctx, baseLog, path); err != nil {
			return err
		}
	}

	return nil
}

func (m *Migrator) applyFile(ctx context.Context, baseLog *logger.Logger, path string) error {
	scopedLog := baseLog
	if scopedLog == nil {
		scopedLog = m.baseLogger()
	}
	scopedLog = scopedLog.With(
		slog.String("file", filepath.Base(path)),
		slog.String("path", path),
	)

	scopedLog.Info("applying migration")

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read migration %q: %w", path, err)
	}

	statement := strings.TrimSpace(string(data))
	if len(statement) == 0 {
		scopedLog.Warn("migration is empty, skipping")
		return nil
	}

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction for migration %q: %w", path, err)
	}

	if _, execErr := tx.ExecContext(ctx, statement); execErr != nil {
		if rbErr := tx.Rollback(); rbErr != nil && rbErr != sql.ErrTxDone {
			scopedLog.Error("rollback error", logger.Err(rbErr))
		}
		return fmt.Errorf("execute migration %q: %w", path, execErr)
	}

	if commitErr := tx.Commit(); commitErr != nil {
		if rbErr := tx.Rollback(); rbErr != nil && rbErr != sql.ErrTxDone {
			scopedLog.Error("rollback error", logger.Err(rbErr))
		}
		return fmt.Errorf("commit migration %q: %w", path, commitErr)
	}

	return nil
}

func isUpMigration(name string) bool {
	return strings.HasSuffix(name, ".up.sql")
}

// ListMigrations returns all .up.sql files in dir in lexical order.
// Useful for debugging and tests.
func ListMigrations(dir fs.FS, root string) ([]string, error) {
	entries, err := fs.ReadDir(dir, root)
	if err != nil {
		return nil, err
	}

	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if isUpMigration(e.Name()) {
			names = append(names, e.Name())
		}
	}

	sort.Strings(names)

	return names, nil
}
