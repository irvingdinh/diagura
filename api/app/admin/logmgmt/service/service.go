package service

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"localhost/app/core/config"
)

var levelRank = map[string]int{
	"DEBUG":   0,
	"INFO":    1,
	"WARN":    2,
	"WARNING": 2,
	"ERROR":   3,
}

// ListFilter holds the filtering and pagination parameters for log queries.
type ListFilter struct {
	Date    string // "2006-01-02" format
	Level   string // "", "DEBUG", "INFO", "WARN", "ERROR"
	Search  string
	Page    int
	PerPage int
}

// ListResult holds a paginated slice of log entries and the total match count.
type ListResult struct {
	Entries []map[string]any
	Total   int
}

// Service provides log file reading and filtering.
type Service struct {
	logsDir string
}

// NewService creates a Service that reads from the configured logs directory.
func NewService() *Service {
	return &Service{
		logsDir: filepath.Join(config.GetString("data_dir"), "logs"),
	}
}

// AvailableDates returns all dates that have log files, sorted newest first.
// Each date is formatted as "2006-01-02".
func (s *Service) AvailableDates(_ context.Context) ([]string, error) {
	entries, err := os.ReadDir(s.logsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("logmgmt: read log directory: %w", err)
	}

	var dates []time.Time
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".log") {
			continue
		}
		name := strings.TrimSuffix(e.Name(), ".log")
		t, err := time.Parse("2006_01_02", name)
		if err != nil {
			continue
		}
		dates = append(dates, t)
	}

	sort.Slice(dates, func(i, j int) bool { return dates[i].After(dates[j]) })

	result := make([]string, len(dates))
	for i, t := range dates {
		result[i] = t.Format("2006-01-02")
	}
	return result, nil
}

// ListEntries reads the log file for the given date, applies filters, and
// returns a paginated result in reverse chronological order (newest first).
func (s *Service) ListEntries(_ context.Context, filter ListFilter) (*ListResult, error) {
	dateFile := strings.ReplaceAll(filter.Date, "-", "_")
	path := filepath.Join(s.logsDir, dateFile+".log")

	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &ListResult{}, nil
		}
		return nil, fmt.Errorf("logmgmt: open log file %s: %w", dateFile, err)
	}
	defer func() { _ = f.Close() }()

	searchLower := strings.ToLower(filter.Search)

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)

	var matched []map[string]any

	for scanner.Scan() {
		line := scanner.Bytes()

		if filter.Search != "" && !strings.Contains(strings.ToLower(string(line)), searchLower) {
			continue
		}

		var entry map[string]any
		if err := json.Unmarshal(line, &entry); err != nil {
			continue
		}

		if filter.Level != "" {
			entryLevel, ok := entry["level"].(string)
			if ok && !meetsMinLevel(entryLevel, filter.Level) {
				continue
			}
		}

		matched = append(matched, entry)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("logmgmt: scan log file %s: %w", dateFile, err)
	}

	total := len(matched)

	// Reverse for newest-first ordering.
	for i, j := 0, total-1; i < j; i, j = i+1, j-1 {
		matched[i], matched[j] = matched[j], matched[i]
	}

	start := (filter.Page - 1) * filter.PerPage
	if start >= total {
		return &ListResult{Entries: []map[string]any{}, Total: total}, nil
	}
	end := start + filter.PerPage
	if end > total {
		end = total
	}

	return &ListResult{
		Entries: matched[start:end],
		Total:   total,
	}, nil
}

func meetsMinLevel(entryLevel, minLevel string) bool {
	if minLevel == "" {
		return true
	}
	return levelRank[strings.ToUpper(entryLevel)] >= levelRank[strings.ToUpper(minLevel)]
}
