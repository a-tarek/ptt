package envfile

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Entry represents a single line in an env file.
type Entry struct {
	Key       string // empty for comments/blank lines
	Value     string
	RawLine   string
	HasExport bool
	QuoteChar byte // 0, '"', or '\''
}

// File represents a parsed env file preserving all formatting.
type File struct {
	Entries []Entry
}

// Parse reads an env file and returns a File preserving all lines.
func Parse(path string) (*File, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open env file: %w", err)
	}
	defer f.Close()

	var entries []Entry
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		entries = append(entries, parseLine(line))
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read env file: %w", err)
	}
	return &File{Entries: entries}, nil
}

func parseLine(line string) Entry {
	trimmed := strings.TrimSpace(line)

	// Blank lines and comments
	if trimmed == "" || strings.HasPrefix(trimmed, "#") {
		return Entry{RawLine: line}
	}

	working := trimmed
	hasExport := false
	if strings.HasPrefix(working, "export ") {
		hasExport = true
		working = strings.TrimPrefix(working, "export ")
		working = strings.TrimSpace(working)
	}

	eqIdx := strings.Index(working, "=")
	if eqIdx < 0 {
		// Not a valid KEY=VALUE line, preserve as raw
		return Entry{RawLine: line}
	}

	key := working[:eqIdx]
	val := working[eqIdx+1:]

	var quoteChar byte
	if len(val) >= 2 {
		first, last := val[0], val[len(val)-1]
		if (first == '"' && last == '"') || (first == '\'' && last == '\'') {
			quoteChar = first
			val = val[1 : len(val)-1]
		}
	}

	return Entry{
		Key:       key,
		Value:     val,
		RawLine:   line,
		HasExport: hasExport,
		QuoteChar: quoteChar,
	}
}

// Get returns the value of a key and whether it was found.
func (f *File) Get(key string) (string, bool) {
	for _, e := range f.Entries {
		if e.Key == key {
			return e.Value, true
		}
	}
	return "", false
}

// Set updates the value of an existing key. Returns error if key not found.
func (f *File) Set(key, value string) error {
	for i, e := range f.Entries {
		if e.Key == key {
			f.Entries[i].Value = value
			f.Entries[i].RawLine = buildLine(e, value)
			return nil
		}
	}
	return fmt.Errorf("key not found: %s", key)
}

func buildLine(e Entry, value string) string {
	var b strings.Builder
	if e.HasExport {
		b.WriteString("export ")
	}
	b.WriteString(e.Key)
	b.WriteByte('=')
	if e.QuoteChar != 0 {
		b.WriteByte(e.QuoteChar)
		b.WriteString(value)
		b.WriteByte(e.QuoteChar)
	} else {
		b.WriteString(value)
	}
	return b.String()
}

// Write writes the env file to the given path, preserving formatting.
func (f *File) Write(path string) error {
	var b strings.Builder
	for i, e := range f.Entries {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(e.RawLine)
	}
	b.WriteByte('\n')
	return os.WriteFile(path, []byte(b.String()), 0644)
}
