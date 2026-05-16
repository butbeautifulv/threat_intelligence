package audit

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// PostgresStore persists audit events in PostgreSQL.
type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(dsn string) (*PostgresStore, error) {
	if dsn == "" {
		return nil, fmt.Errorf("postgres dsn required")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	s := &PostgresStore{db: db}
	if err := s.ensureSchema(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

func (s *PostgresStore) ensureSchema(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS engage_audit_events (
  id BIGSERIAL PRIMARY KEY,
  subject TEXT NOT NULL DEFAULT '',
  tool TEXT NOT NULL,
  target TEXT NOT NULL DEFAULT '',
  job_id TEXT NOT NULL DEFAULT '',
  success BOOLEAN NOT NULL,
  error TEXT NOT NULL DEFAULT '',
  at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_engage_audit_at ON engage_audit_events(at DESC);
`)
	return err
}

func (s *PostgresStore) Append(e Event) error {
	_, err := s.db.Exec(
		`INSERT INTO engage_audit_events (subject, tool, target, job_id, success, error, at) VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		e.Subject, e.Tool, e.Target, e.JobID, e.Success, e.Error, e.At.UTC(),
	)
	return err
}

func (s *PostgresStore) Recent(limit int) ([]Event, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := s.db.Query(
		`SELECT subject, tool, target, job_id, success, error, at FROM engage_audit_events ORDER BY at DESC LIMIT $1`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Event
	for rows.Next() {
		var e Event
		if err := rows.Scan(&e.Subject, &e.Tool, &e.Target, &e.JobID, &e.Success, &e.Error, &e.At); err != nil {
			continue
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// Retention deletes events older than the given duration.
func (s *PostgresStore) Retention(olderThan time.Duration) (int64, error) {
	cutoff := time.Now().UTC().Add(-olderThan)
	res, err := s.db.Exec(`DELETE FROM engage_audit_events WHERE at < $1`, cutoff)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (s *PostgresStore) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// ExportNDJSON returns events since the given time (zero = all), one JSON object per line.
func (s *PostgresStore) ExportNDJSON(since time.Time) ([]byte, error) {
	var rows *sql.Rows
	var err error
	if since.IsZero() {
		rows, err = s.db.Query(
			`SELECT subject, tool, target, job_id, success, error, at FROM engage_audit_events ORDER BY at ASC`,
		)
	} else {
		rows, err = s.db.Query(
			`SELECT subject, tool, target, job_id, success, error, at FROM engage_audit_events WHERE at >= $1 ORDER BY at ASC`,
			since.UTC(),
		)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var buf bytes.Buffer
	for rows.Next() {
		var e Event
		if err := rows.Scan(&e.Subject, &e.Tool, &e.Target, &e.JobID, &e.Success, &e.Error, &e.At); err != nil {
			continue
		}
		b, err := json.Marshal(e)
		if err != nil {
			continue
		}
		buf.Write(b)
		buf.WriteByte('\n')
	}
	return buf.Bytes(), rows.Err()
}
