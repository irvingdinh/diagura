package migration

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantUp   []string
		wantDown []string
	}{
		{
			name: "basic up and down",
			input: `-- +migration Up
CREATE TABLE a (id TEXT);
CREATE TABLE b (id TEXT);

-- +migration Down
DROP TABLE b;
DROP TABLE a;
`,
			wantUp:   []string{"CREATE TABLE a (id TEXT)", "CREATE TABLE b (id TEXT)"},
			wantDown: []string{"DROP TABLE b", "DROP TABLE a"},
		},
		{
			name: "semicolon inside single-quoted string",
			input: `-- +migration Up
INSERT INTO t (val) VALUES ('hello;world');
`,
			wantUp: []string{"INSERT INTO t (val) VALUES ('hello;world')"},
		},
		{
			name: "escaped quote with semicolon",
			input: `-- +migration Up
INSERT INTO t (val) VALUES ('it''s ; here');
`,
			wantUp: []string{"INSERT INTO t (val) VALUES ('it''s ; here')"},
		},
		{
			name: "multiple semicolons on one line",
			input: `-- +migration Up
INSERT INTO a VALUES (1); INSERT INTO b VALUES (2);
`,
			wantUp: []string{"INSERT INTO a VALUES (1)", "INSERT INTO b VALUES (2)"},
		},
		{
			name: "mixed quoted and unquoted semicolons",
			input: `-- +migration Up
INSERT INTO a VALUES ('x;y'); INSERT INTO b VALUES ('p;q');
`,
			wantUp: []string{"INSERT INTO a VALUES ('x;y')", "INSERT INTO b VALUES ('p;q')"},
		},
		{
			name: "StatementBegin and StatementEnd",
			input: `-- +migration Up
-- +migration StatementBegin
CREATE TRIGGER t AFTER INSERT ON a
BEGIN
  UPDATE b SET n = n + 1;
END;
-- +migration StatementEnd
`,
			wantUp: []string{"CREATE TRIGGER t AFTER INSERT ON a\nBEGIN\n  UPDATE b SET n = n + 1;\nEND;"},
		},
		{
			name: "no down section",
			input: `-- +migration Up
CREATE TABLE t (id TEXT);
`,
			wantUp:   []string{"CREATE TABLE t (id TEXT)"},
			wantDown: nil,
		},
		{
			name: "content before first annotation is ignored",
			input: `-- This is a comment
-- +migration Up
CREATE TABLE t (id TEXT);
`,
			wantUp: []string{"CREATE TABLE t (id TEXT)"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}

			if !slicesEqual(result.Up, tt.wantUp) {
				t.Errorf("Up:\n  got  %q\n  want %q", result.Up, tt.wantUp)
			}
			if !slicesEqual(result.Down, tt.wantDown) {
				t.Errorf("Down:\n  got  %q\n  want %q", result.Down, tt.wantDown)
			}
		})
	}
}

func slicesEqual(a, b []string) bool {
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
