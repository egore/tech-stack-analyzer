package progress

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// Event types that the scanner reports
type EventType int

const (
	EventScanStart EventType = iota
	EventScanComplete
	EventEnterDirectory
	EventLeaveDirectory
	EventComponentDetected
	EventFileProcessing
	EventSkipped
	EventProgress
)

// Event represents something that happened during scanning
type Event struct {
	Type      EventType
	Path      string
	Name      string
	Tech      string
	Info      string
	Reason    string
	FileCount int
	DirCount  int
	Duration  time.Duration
}

// Reporter is the interface the scanner uses to report events
type Reporter interface {
	Report(event Event)
}

// Handler processes events and produces output
type Handler interface {
	Handle(event Event)
}

// Progress is the centralized verbose system
type Progress struct {
	enabled bool
	handler Handler
}

// New creates a new progress reporter
func New(enabled bool, handler Handler) *Progress {
	if handler == nil {
		handler = NewSimpleHandler(os.Stderr)
	}
	return &Progress{
		enabled: enabled,
		handler: handler,
	}
}

// Report sends an event to the handler (only if enabled)
func (p *Progress) Report(event Event) {
	if !p.enabled {
		return
	}
	p.handler.Handle(event)
}

// Convenience methods for the scanner to report events

func (p *Progress) ScanStart(path string, excludePatterns []string) {
	p.Report(Event{
		Type: EventScanStart,
		Path: path,
		Info: strings.Join(excludePatterns, ", "),
	})
}

func (p *Progress) ScanComplete(files, dirs int, duration time.Duration) {
	p.Report(Event{
		Type:      EventScanComplete,
		FileCount: files,
		DirCount:  dirs,
		Duration:  duration,
	})
}

func (p *Progress) EnterDirectory(path string) {
	p.Report(Event{
		Type: EventEnterDirectory,
		Path: path,
	})
}

func (p *Progress) LeaveDirectory(path string) {
	p.Report(Event{
		Type: EventLeaveDirectory,
		Path: path,
	})
}

func (p *Progress) ComponentDetected(name, tech, path string) {
	p.Report(Event{
		Type: EventComponentDetected,
		Name: name,
		Tech: tech,
		Path: path,
	})
}

func (p *Progress) FileProcessing(path, info string) {
	p.Report(Event{
		Type: EventFileProcessing,
		Path: path,
		Info: info,
	})
}

func (p *Progress) Skipped(path, reason string) {
	p.Report(Event{
		Type:   EventSkipped,
		Path:   path,
		Reason: reason,
	})
}

func (p *Progress) ProgressUpdate(files, dirs int) {
	p.Report(Event{
		Type:      EventProgress,
		FileCount: files,
		DirCount:  dirs,
	})
}

// SimpleHandler outputs events as simple lines (no tree)
type SimpleHandler struct {
	writer io.Writer
}

func NewSimpleHandler(writer io.Writer) *SimpleHandler {
	return &SimpleHandler{writer: writer}
}

func (h *SimpleHandler) Handle(event Event) {
	switch event.Type {
	case EventScanStart:
		fmt.Fprintf(h.writer, "[SCAN] Starting: %s\n", event.Path)
		if event.Info != "" {
			fmt.Fprintf(h.writer, "[SCAN] Excluding: %s\n", event.Info)
		}

	case EventScanComplete:
		fmt.Fprintf(h.writer, "[SCAN] Completed: %d files, %d directories in %.1fs\n",
			event.FileCount, event.DirCount, event.Duration.Seconds())

	case EventEnterDirectory:
		fmt.Fprintf(h.writer, "[DIR]  Entering: %s\n", event.Path)

	case EventLeaveDirectory:
		// Simple handler doesn't show leave events
		// (TreeHandler would use this)

	case EventComponentDetected:
		fmt.Fprintf(h.writer, "[COMP] Detected: %s (%s) at %s\n",
			event.Name, event.Tech, event.Path)

	case EventFileProcessing:
		fmt.Fprintf(h.writer, "[FILE] Parsing: %s (%s)\n", event.Path, event.Info)

	case EventSkipped:
		fmt.Fprintf(h.writer, "[SKIP] Excluding: %s (%s)\n", event.Path, event.Reason)

	case EventProgress:
		fmt.Fprintf(h.writer, "[PROG] Progress: %d files, %d directories\n",
			event.FileCount, event.DirCount)
	}
}

// TreeHandler outputs events with tree-like visualization
type TreeHandler struct {
	writer io.Writer
	depth  int
}

func NewTreeHandler(writer io.Writer) *TreeHandler {
	return &TreeHandler{
		writer: writer,
		depth:  0,
	}
}

func (h *TreeHandler) Handle(event Event) {
	indent := strings.Repeat("│  ", h.depth)
	prefix := "├─ "

	switch event.Type {
	case EventScanStart:
		fmt.Fprintf(h.writer, "Scanning %s...\n", event.Path)
		if event.Info != "" {
			fmt.Fprintf(h.writer, "Excluding: %s\n", event.Info)
		}
		fmt.Fprintln(h.writer)

	case EventScanComplete:
		fmt.Fprintf(h.writer, "└─ Completed: %d files, %d directories in %.1fs\n",
			event.FileCount, event.DirCount, event.Duration.Seconds())

	case EventEnterDirectory:
		fmt.Fprintf(h.writer, "%s%s%s\n", indent, prefix, event.Path)
		h.depth++

	case EventLeaveDirectory:
		h.depth--
		if h.depth < 0 {
			h.depth = 0
		}

	case EventComponentDetected:
		fmt.Fprintf(h.writer, "%s%sDetected: %s (%s)\n",
			indent, prefix, event.Name, event.Tech)

	case EventFileProcessing:
		fmt.Fprintf(h.writer, "%s%sParsing: %s (%s)\n",
			indent, prefix, event.Path, event.Info)

	case EventSkipped:
		fmt.Fprintf(h.writer, "%s%sSkipping: %s (%s)\n",
			indent, prefix, event.Path, event.Reason)

	case EventProgress:
		fmt.Fprintf(h.writer, "%s%sProgress: %d files, %d directories\n",
			indent, prefix, event.FileCount, event.DirCount)
	}
}

// NullHandler discards all events (for disabled verbose mode)
type NullHandler struct{}

func NewNullHandler() *NullHandler {
	return &NullHandler{}
}

func (h *NullHandler) Handle(event Event) {
	// Do nothing
}
