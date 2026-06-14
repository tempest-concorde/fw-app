package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
)

// Writer appends audit events to a JSONL file in CloudEvents v1.0 format.
type Writer struct {
	file *os.File
	mu   sync.Mutex
}

// NewWriter creates a new audit writer that appends events to the specified file.
// The file is created if it doesn't exist, or opened for append if it does.
func NewWriter(path string) (*Writer, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		return nil, fmt.Errorf("failed to create audit log directory: %w", err)
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to open audit file: %w", err)
	}

	return &Writer{
		file: file,
	}, nil
}

// WriteEvent appends a CloudEvents v1.0 audit event to the file.
// Each event is written as a single line of JSON (JSONL format).
//
// Parameters:
//   - ctx: context for the operation
//   - eventType: the CloudEvents type field (e.g., "com.example.flight.created")
//   - source: the CloudEvents source field (e.g., "flight-wall/api")
//   - subject: the CloudEvents subject field (optional, use "" to omit)
//   - data: the event payload (will be JSON-marshaled into the data field)
//   - extensions: optional CloudEvents extensions (e.g., fwuser, fwaction, fwstatus)
func (w *Writer) WriteEvent(ctx context.Context, eventType, source, subject string, data interface{}, extensions map[string]string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	event := cloudevents.NewEvent()
	event.SetID(uuid.New().String())
	event.SetType(eventType)
	event.SetSource(source)
	event.SetTime(time.Now().UTC())

	if subject != "" {
		event.SetSubject(subject)
	}

	if err := event.SetData(cloudevents.ApplicationJSON, data); err != nil {
		return fmt.Errorf("failed to set event data: %w", err)
	}

	// Add extensions
	for key, value := range extensions {
		event.SetExtension(key, value)
	}

	// Marshal to JSON
	eventBytes, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Append newline for JSONL format
	eventBytes = append(eventBytes, '\n')

	// Write to file
	if _, err := w.file.Write(eventBytes); err != nil {
		return fmt.Errorf("failed to write event: %w", err)
	}

	return nil
}

// Close closes the underlying file.
func (w *Writer) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file != nil {
		return w.file.Close()
	}
	return nil
}
