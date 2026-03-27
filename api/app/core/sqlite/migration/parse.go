package migration

import (
	"bufio"
	"io"
	"strings"
)

const (
	annotationUp             = "-- +migration Up"
	annotationDown           = "-- +migration Down"
	annotationStatementBegin = "-- +migration StatementBegin"
	annotationStatementEnd   = "-- +migration StatementEnd"
)

type parsedMigration struct {
	Up   []string
	Down []string
}

func parse(r io.Reader) (parsedMigration, error) {
	var result parsedMigration
	scanner := bufio.NewScanner(r)

	var (
		buf         strings.Builder
		direction   string // "up" or "down"
		inStatement bool   // inside StatementBegin..StatementEnd
	)

	flush := func() {
		s := strings.TrimSpace(buf.String())
		buf.Reset()
		if s == "" {
			return
		}
		switch direction {
		case "up":
			result.Up = append(result.Up, s)
		case "down":
			result.Down = append(result.Down, s)
		}
	}

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		switch trimmed {
		case annotationUp:
			flush()
			direction = "up"
			continue
		case annotationDown:
			flush()
			direction = "down"
			continue
		case annotationStatementBegin:
			inStatement = true
			continue
		case annotationStatementEnd:
			inStatement = false
			flush()
			continue
		}

		if direction == "" {
			continue
		}

		if inStatement {
			if buf.Len() > 0 {
				buf.WriteByte('\n')
			}
			buf.WriteString(line)
			continue
		}

		// Split on semicolons outside single-quoted strings.
		inQuote := false
		start := 0
		for i := 0; i < len(line); i++ {
			ch := line[i]
			if ch == '\'' {
				inQuote = !inQuote
			} else if ch == ';' && !inQuote {
				if buf.Len() > 0 {
					buf.WriteByte('\n')
				}
				buf.WriteString(line[start:i])
				flush()
				start = i + 1
			}
		}
		remaining := line[start:]
		if len(remaining) > 0 {
			if buf.Len() > 0 {
				buf.WriteByte('\n')
			}
			buf.WriteString(remaining)
		}
	}
	flush()

	return result, scanner.Err()
}
