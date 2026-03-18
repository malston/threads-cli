package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

type samplePost struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

func TestPrintJSON(t *testing.T) {
	var buf bytes.Buffer
	post := samplePost{ID: "123", Text: "hello world"}

	if err := Print(&buf, post, JSON); err != nil {
		t.Fatalf("Print returned error: %v", err)
	}

	got := buf.String()

	// Must be valid JSON
	var decoded map[string]any
	if err := json.Unmarshal([]byte(got), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, got)
	}

	// Must be indented (contains newlines and spaces)
	if !strings.Contains(got, "\n") {
		t.Errorf("expected indented JSON with newlines, got: %s", got)
	}
	if !strings.Contains(got, "  ") {
		t.Errorf("expected 2-space indent, got: %s", got)
	}

	// Check values
	if decoded["id"] != "123" {
		t.Errorf("expected id=123, got %v", decoded["id"])
	}
	if decoded["text"] != "hello world" {
		t.Errorf("expected text=hello world, got %v", decoded["text"])
	}
}

func TestPrintJSONSlice(t *testing.T) {
	var buf bytes.Buffer
	posts := []samplePost{
		{ID: "1", Text: "first"},
		{ID: "2", Text: "second"},
	}

	if err := Print(&buf, posts, JSON); err != nil {
		t.Fatalf("Print returned error: %v", err)
	}

	var decoded []map[string]any
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON array: %v\noutput: %s", err, buf.String())
	}

	if len(decoded) != 2 {
		t.Fatalf("expected 2 items, got %d", len(decoded))
	}
}

func TestPrintTextStruct(t *testing.T) {
	var buf bytes.Buffer
	post := samplePost{ID: "456", Text: "some text"}

	if err := Print(&buf, post, Text); err != nil {
		t.Fatalf("Print returned error: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "ID: 456") {
		t.Errorf("expected 'ID: 456' in output, got: %s", got)
	}
	if !strings.Contains(got, "Text: some text") {
		t.Errorf("expected 'Text: some text' in output, got: %s", got)
	}
}

func TestPrintTextMap(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]string{
		"name": "alice",
	}

	if err := Print(&buf, data, Text); err != nil {
		t.Fatalf("Print returned error: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "name: alice") {
		t.Errorf("expected 'name: alice' in output, got: %s", got)
	}
}

func TestPrintTextSlice(t *testing.T) {
	var buf bytes.Buffer
	items := []string{"one", "two", "three"}

	if err := Print(&buf, items, Text); err != nil {
		t.Fatalf("Print returned error: %v", err)
	}

	got := buf.String()
	lines := strings.Split(strings.TrimSpace(got), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d: %s", len(lines), got)
	}
	if lines[0] != "one" {
		t.Errorf("expected first line 'one', got %q", lines[0])
	}
}

func TestPrintTableSliceOfStructs(t *testing.T) {
	var buf bytes.Buffer
	posts := []samplePost{
		{ID: "1", Text: "first post"},
		{ID: "2", Text: "second post"},
	}

	if err := Print(&buf, posts, Table); err != nil {
		t.Fatalf("Print returned error: %v", err)
	}

	got := buf.String()
	lines := strings.Split(strings.TrimSpace(got), "\n")

	if len(lines) != 3 {
		t.Fatalf("expected 3 lines (header + 2 data), got %d:\n%s", len(lines), got)
	}

	// Header must contain field names
	header := lines[0]
	if !strings.Contains(header, "ID") || !strings.Contains(header, "Text") {
		t.Errorf("header missing field names, got: %s", header)
	}

	// Data rows must contain values
	if !strings.Contains(lines[1], "1") || !strings.Contains(lines[1], "first post") {
		t.Errorf("first data row missing values, got: %s", lines[1])
	}
	if !strings.Contains(lines[2], "2") || !strings.Contains(lines[2], "second post") {
		t.Errorf("second data row missing values, got: %s", lines[2])
	}
}

func TestPrintTableSliceOfMaps(t *testing.T) {
	var buf bytes.Buffer
	data := []map[string]string{
		{"id": "10", "name": "alice"},
		{"id": "20", "name": "bob"},
	}

	if err := Print(&buf, data, Table); err != nil {
		t.Fatalf("Print returned error: %v", err)
	}

	got := buf.String()
	lines := strings.Split(strings.TrimSpace(got), "\n")

	if len(lines) != 3 {
		t.Fatalf("expected 3 lines (header + 2 data), got %d:\n%s", len(lines), got)
	}
}

func TestParseFormatValid(t *testing.T) {
	tests := []struct {
		input string
		want  Format
	}{
		{"json", JSON},
		{"text", Text},
		{"table", Table},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseFormat(tt.input)
			if err != nil {
				t.Fatalf("ParseFormat(%q) returned error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("ParseFormat(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseFormatInvalid(t *testing.T) {
	invalids := []string{"xml", "yaml", "csv", "", "JSON"}

	for _, s := range invalids {
		t.Run(s, func(t *testing.T) {
			_, err := ParseFormat(s)
			if err == nil {
				t.Errorf("ParseFormat(%q) expected error, got nil", s)
			}
		})
	}
}
