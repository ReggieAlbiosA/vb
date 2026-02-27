package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Entry is a single timestamped usage log line.
type Entry struct {
	Timestamp time.Time
	Topic     string
	Lens      string
}

// Append writes one Entry to USED.md inside topicDir.
// File is created if it does not exist. Existing content is never overwritten.
func Append(topicDir, topic, lens string) error {
	usedPath := filepath.Join(topicDir, "USED.md")

	entry := formatEntry(Entry{
		Timestamp: time.Now().UTC(),
		Topic:     topic,
		Lens:      lens,
	})

	f, err := os.OpenFile(usedPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("opening USED.md: %w", err)
	}
	defer f.Close()

	_, err = fmt.Fprintln(f, entry)
	return err
}

// formatEntry produces a single, human-readable log line.
func formatEntry(e Entry) string {
	return fmt.Sprintf("- %s  vb %s --%s",
		e.Timestamp.Format("2006-01-02 15:04 UTC"),
		e.Topic,
		e.Lens,
	)
}
