package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/seanb4t/fastmail-cli/internal/jmap"
)

func TestFormatJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		want    string
		wantErr bool
	}{
		{
			name:  "simple struct",
			input: struct{ Name string }{Name: "test"},
			want:  "{\n  \"Name\": \"test\"\n}",
		},
		{
			name:  "map",
			input: map[string]int{"count": 42},
			want:  "{\n  \"count\": 42\n}",
		},
		{
			name:  "slice",
			input: []string{"a", "b"},
			want:  "[\n  \"a\",\n  \"b\"\n]",
		},
		{
			name:  "nil",
			input: nil,
			want:  "null",
		},
		{
			name:    "channel (unmarshalable)",
			input:   make(chan int),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FormatJSON(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFormatTable(t *testing.T) {
	tests := []struct {
		name    string
		headers []string
		rows    [][]string
		want    string
	}{
		{
			name:    "empty headers",
			headers: []string{},
			rows:    [][]string{{"a", "b"}},
			want:    "",
		},
		{
			name:    "simple table",
			headers: []string{"ID", "Name"},
			rows: [][]string{
				{"1", "Alice"},
				{"2", "Bob"},
			},
			want: "ID  Name \n--  -----\n1   Alice\n2   Bob  \n",
		},
		{
			name:    "varying column widths",
			headers: []string{"A", "Longer Header"},
			rows: [][]string{
				{"short", "x"},
				{"a", "much longer value"},
			},
			want: "A      Longer Header    \n-----  -----------------\nshort  x                \na      much longer value\n",
		},
		{
			name:    "empty rows",
			headers: []string{"Col1", "Col2"},
			rows:    [][]string{},
			want:    "Col1  Col2\n----  ----\n",
		},
		{
			name:    "missing cells in row",
			headers: []string{"A", "B", "C"},
			rows: [][]string{
				{"1"},
				{"x", "y", "z"},
			},
			want: "A  B  C\n-  -  -\n1      \nx  y  z\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatTable(tt.headers, tt.rows)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestWriter_PrintEmail_JSON(t *testing.T) {
	var buf bytes.Buffer
	w := New(WithJSON(true), WithOutput(&buf, &buf))

	email := &jmap.Email{
		ID:         "M123",
		Subject:    "Test Subject",
		ReceivedAt: "2024-01-15T10:30:00Z",
		Preview:    "This is a preview",
	}

	w.PrintEmail(email)

	var result jmap.Email
	err := json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, email.ID, result.ID)
	assert.Equal(t, email.Subject, result.Subject)
}

func TestWriter_PrintEmail_Quiet(t *testing.T) {
	var buf bytes.Buffer
	w := New(WithQuiet(true), WithOutput(&buf, &buf))

	email := &jmap.Email{
		ID:      "M123",
		Subject: "Test Subject",
	}

	w.PrintEmail(email)

	assert.Empty(t, buf.String())
}

func TestWriter_PrintEmailList_JSON(t *testing.T) {
	var buf bytes.Buffer
	w := New(WithJSON(true), WithOutput(&buf, &buf))

	emails := []jmap.Email{
		{ID: "M1", Subject: "First"},
		{ID: "M2", Subject: "Second"},
	}

	w.PrintEmailList(emails)

	var result []jmap.Email
	err := json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "M1", result[0].ID)
}

func TestWriter_PrintEmailList_Quiet(t *testing.T) {
	var buf bytes.Buffer
	w := New(WithQuiet(true), WithOutput(&buf, &buf))

	emails := []jmap.Email{
		{ID: "M1", Subject: "First"},
	}

	w.PrintEmailList(emails)

	assert.Empty(t, buf.String())
}

func TestWriter_Error(t *testing.T) {
	var stdout, stderr bytes.Buffer
	w := New(WithOutput(&stdout, &stderr))

	w.Error("something went %s", "wrong")

	assert.Empty(t, stdout.String())
	assert.Equal(t, "error: something went wrong\n", stderr.String())
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input  string
		maxLen int
		want   string
	}{
		{"short", 10, "short"},
		{"exactly10!", 10, "exactly10!"},
		{"this is too long", 10, "this is..."},
		{"ab", 3, "ab"},
		{"abcd", 3, "abc"},
		{"", 5, ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := truncate(tt.input, tt.maxLen)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestWriter_Print(t *testing.T) {
	var buf bytes.Buffer
	w := New(WithJSON(true), WithOutput(&buf, &buf))

	data := map[string]string{"key": "value"}
	w.Print(data)

	assert.True(t, strings.Contains(buf.String(), `"key": "value"`))
}
