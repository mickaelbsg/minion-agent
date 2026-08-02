package storage

import (
	"database/sql"
	"errors"
	"fmt"
)

var ErrIdempotencyPayloadMismatch = errors.New("request id already used with a different payload")

type IdempotencyState string

const (
	IdempotencyClaimed    IdempotencyState = "claimed"
	IdempotencyInProgress IdempotencyState = "in_progress"
	IdempotencyCompleted  IdempotencyState = "completed"
)

type IdempotencyRecord struct {
	State        IdempotencyState
	StatusCode   int
	ResponseBody []byte
}

func (s *Storage) ensureIdempotencySchema() error {
	if s == nil || s.DB == nil {
		return errors.New("storage unavailable")
	}
	_, err := s.DB.Exec(`CREATE TABLE IF NOT EXISTS idempotency_records (
		client_name TEXT NOT NULL,
		action TEXT NOT NULL,
		request_id TEXT NOT NULL,
		payload_hash TEXT NOT NULL,
		state TEXT NOT NULL,
		status_code INTEGER,
		response_body BLOB,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (client_name, action, request_id)
	);`)
	return err
}

func (s *Storage) ClaimIdempotency(clientName, action, requestID, payloadHash string) (IdempotencyRecord, error) {
	if err := s.ensureIdempotencySchema(); err != nil {
		return IdempotencyRecord{}, err
	}

	result, err := s.DB.Exec(
		`INSERT OR IGNORE INTO idempotency_records
		(client_name, action, request_id, payload_hash, state)
		VALUES (?, ?, ?, ?, ?)`,
		clientName, action, requestID, payloadHash, string(IdempotencyInProgress),
	)
	if err != nil {
		return IdempotencyRecord{}, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return IdempotencyRecord{}, err
	}
	if rows == 1 {
		return IdempotencyRecord{State: IdempotencyClaimed}, nil
	}

	var storedHash, state string
	var status sql.NullInt64
	var response []byte
	err = s.DB.QueryRow(
		`SELECT payload_hash, state, status_code, response_body
		FROM idempotency_records
		WHERE client_name = ? AND action = ? AND request_id = ?`,
		clientName, action, requestID,
	).Scan(&storedHash, &state, &status, &response)
	if err != nil {
		return IdempotencyRecord{}, err
	}
	if storedHash != payloadHash {
		return IdempotencyRecord{}, ErrIdempotencyPayloadMismatch
	}

	record := IdempotencyRecord{State: IdempotencyState(state), ResponseBody: response}
	if status.Valid {
		record.StatusCode = int(status.Int64)
	}
	return record, nil
}

func (s *Storage) CompleteIdempotency(clientName, action, requestID string, statusCode int, responseBody []byte) error {
	if err := s.ensureIdempotencySchema(); err != nil {
		return err
	}
	result, err := s.DB.Exec(
		`UPDATE idempotency_records
		SET state = ?, status_code = ?, response_body = ?, updated_at = CURRENT_TIMESTAMP
		WHERE client_name = ? AND action = ? AND request_id = ? AND state = ?`,
		string(IdempotencyCompleted), statusCode, responseBody,
		clientName, action, requestID, string(IdempotencyInProgress),
	)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return fmt.Errorf("idempotency record not found or already completed")
	}
	return nil
}
