package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

var dbConn *sql.DB

// InitDB initializes the SQLite database.
func InitDB() error {
	cacheDir, err := GetCacheDir()
	if err != nil {
		return fmt.Errorf("get cache dir: %w", err)
	}
	// Database path under ~/.cache/teams-tui-go/
	dbPath := filepath.Join(cacheDir, "teams-tui-go.db")

	if err := os.MkdirAll(filepath.Dir(dbPath), 0o700); err != nil {
		return fmt.Errorf("create cache directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("open sqlite db: %w", err)
	}

	// Create tables if they do not exist.
	queries := []string{
		`CREATE TABLE IF NOT EXISTS messages (
			id TEXT PRIMARY KEY,
			conversation_id TEXT NOT NULL,
			created_date_time TEXT NOT NULL,
			is_reply INTEGER DEFAULT 0,
			reply_to_id TEXT,
			data TEXT NOT NULL
		);`,
		`CREATE INDEX IF NOT EXISTS idx_messages_conv_created ON messages(conversation_id, created_date_time DESC);`,
		`CREATE TABLE IF NOT EXISTS conversations (
			id TEXT PRIMARY KEY,
			next_link TEXT
		);`,
	}

	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			_ = db.Close()
			return fmt.Errorf("init db query failed: %w", err)
		}
	}

	dbConn = db
	return nil
}

// CloseDB closes the database connection.
func CloseDB() {
	if dbConn != nil {
		_ = dbConn.Close()
	}
}

// SaveMessages stores a list of messages.
func SaveMessages(conversationID string, msgs []Message) error {
	if dbConn == nil {
		return nil
	}

	tx, err := dbConn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `INSERT INTO messages (id, conversation_id, created_date_time, is_reply, reply_to_id, data)
              VALUES (?, ?, ?, ?, ?, ?)
              ON CONFLICT(id) DO UPDATE SET
                created_date_time = excluded.created_date_time,
                is_reply = excluded.is_reply,
                reply_to_id = excluded.reply_to_id,
                data = excluded.data;`

	stmt, err := tx.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, msg := range msgs {
		dataBytes, err := json.Marshal(msg)
		if err != nil {
			return err
		}
		isReplyVal := 0
		if msg.IsReply {
			isReplyVal = 1
		}
		_, err = stmt.Exec(msg.ID, conversationID, msg.CreatedDateTime, isReplyVal, msg.ReplyToID, string(dataBytes))
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetStoredMessages retrieves stored messages for a conversation, newest first.
func GetStoredMessages(conversationID string, limit int) ([]Message, error) {
	if dbConn == nil {
		return nil, nil
	}

	query := `SELECT data FROM messages WHERE conversation_id = ? ORDER BY created_date_time DESC LIMIT ?`
	rows, err := dbConn.Query(query, conversationID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []Message
	for rows.Next() {
		var dataStr string
		if err := rows.Scan(&dataStr); err != nil {
			return nil, err
		}
		var msg Message
		if err := json.Unmarshal([]byte(dataStr), &msg); err != nil {
			return nil, err
		}
		msgs = append(msgs, msg)
	}
	sortMessagesNewestFirst(msgs)

	return msgs, nil
}

// SaveNextLink stores the next link for a conversation.
func SaveNextLink(conversationID string, nextLink string) error {
	if dbConn == nil {
		return nil
	}

	query := `INSERT INTO conversations (id, next_link) VALUES (?, ?)
              ON CONFLICT(id) DO UPDATE SET next_link = excluded.next_link;`
	_, err := dbConn.Exec(query, conversationID, nextLink)
	return err
}

// GetNextLink retrieves the next link for a conversation.
func GetNextLink(conversationID string) (string, error) {
	if dbConn == nil {
		return "", nil
	}

	query := `SELECT next_link FROM conversations WHERE id = ?`
	var nextLink string
	err := dbConn.QueryRow(query, conversationID).Scan(&nextLink)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return nextLink, err
}

// UpdateStoredMessageBody updates the body content of a stored message.
func UpdateStoredMessageBody(messageID string, newContent string) error {
	if dbConn == nil {
		return nil
	}
	// Fetch the message first
	var dataStr string
	querySelect := `SELECT data FROM messages WHERE id = ?`
	err := dbConn.QueryRow(querySelect, messageID).Scan(&dataStr)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}

	var msg Message
	if err := json.Unmarshal([]byte(dataStr), &msg); err != nil {
		return err
	}

	if msg.Body == nil {
		msg.Body = &MessageBody{}
	}
	msg.Body.Content = &newContent
	msg.PlainTextCached = nil // force re-render

	dataBytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	queryUpdate := `UPDATE messages SET data = ? WHERE id = ?`
	_, err = dbConn.Exec(queryUpdate, string(dataBytes), messageID)
	return err
}
