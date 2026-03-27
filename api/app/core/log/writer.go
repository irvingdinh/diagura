package log

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// dailyFileWriter is the file sink behind the JSON file handler: a
// thread-safe io.WriteCloser that appends bytes to one JSONL file per calendar
// day (YYYY_MM_DD.log) under a fixed directory. Writes are buffered (64 KB)
// and flushed periodically (every 200 ms) to batch syscalls while keeping log
// entries visible with low latency.
//
// Rotation happens on the first write after the date changes; the mutex
// serializes writes and protects the current file handle, buffer, and date.
type dailyFileWriter struct {
	mu    sync.Mutex
	dir   string
	date  string
	file  *os.File
	buf   *bufio.Writer
	nowFn func() time.Time
	done  chan struct{}
}

func newDailyFileWriter(dir string) *dailyFileWriter {
	w := &dailyFileWriter{
		dir:   dir,
		nowFn: time.Now,
		done:  make(chan struct{}),
	}
	go w.flushLoop()
	return w
}

func (w *dailyFileWriter) flushLoop() {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			w.mu.Lock()
			if w.buf != nil {
				_ = w.buf.Flush()
			}
			w.mu.Unlock()
		case <-w.done:
			return
		}
	}
}

// Write appends p to today's log file, rotating to a new file when the
// calendar day changes.
func (w *dailyFileWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	today := w.nowFn().Format("2006_01_02")
	if today != w.date {
		if err := w.rotate(today); err != nil {
			return 0, fmt.Errorf("log: rotate: %w", err)
		}
	}
	return w.buf.Write(p)
}

func (w *dailyFileWriter) rotate(date string) error {
	if w.buf != nil {
		_ = w.buf.Flush()
	}
	if w.file != nil {
		_ = w.file.Close()
	}
	path := filepath.Join(w.dir, date+".log")
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	w.file = f
	w.buf = bufio.NewWriterSize(f, 64*1024)
	w.date = date
	return nil
}

// Flush writes any buffered data to the underlying file.
func (w *dailyFileWriter) Flush() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.buf != nil {
		return w.buf.Flush()
	}
	return nil
}

// Close stops the background flush goroutine, flushes remaining data, and
// closes the file handle. Safe to call multiple times.
func (w *dailyFileWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	select {
	case <-w.done:
	default:
		close(w.done)
	}

	if w.buf != nil {
		_ = w.buf.Flush()
		w.buf = nil
	}
	if w.file != nil {
		err := w.file.Close()
		w.file = nil
		w.date = ""
		return err
	}
	return nil
}
