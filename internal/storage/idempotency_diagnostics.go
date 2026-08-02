package storage

import (
	"errors"
	"time"
)

const MaxIdempotencyDiagnosticsLimit = 100
const maxIdempotencyDiagnosticsQueryLimit = MaxIdempotencyDiagnosticsLimit + 1

type IdempotencyDiagnostic struct {
	ClientName string    `json:"client_name"`
	Action     string    `json:"action"`
	RequestID  string    `json:"request_id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (s *Storage) ListInProgressIdempotency(action string, limit, offset int) ([]IdempotencyDiagnostic, error) {
	if err := s.ensureIdempotencySchema(); err != nil {
		return nil, err
	}
	if limit < 1 || limit > maxIdempotencyDiagnosticsQueryLimit {
		return nil, errors.New("idempotency diagnostics query limit must be between 1 and 101")
	}
	if offset < 0 {
		return nil, errors.New("idempotency diagnostics offset must not be negative")
	}

	query := `SELECT client_name, action, request_id, created_at, updated_at
		FROM idempotency_records
		WHERE state = ?`
	args := []interface{}{string(IdempotencyInProgress)}
	if action != "" {
		query += ` AND action = ?`
		args = append(args, action)
	}
	query += ` ORDER BY created_at ASC, request_id ASC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := s.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]IdempotencyDiagnostic, 0)
	for rows.Next() {
		var item IdempotencyDiagnostic
		if err := rows.Scan(&item.ClientName, &item.Action, &item.RequestID, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
