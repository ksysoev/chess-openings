package openings

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseOpenings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     string
		wantCount int
		wantErr   bool
	}{
		{
			name:      "valid TSV with header",
			input:     "eco\tname\tpgn\nA00\tAmar Opening\t1. Nh3\n",
			wantCount: 1,
		},
		{
			name:      "multiple entries",
			input:     "eco\tname\tpgn\nA00\tAmar Opening\t1. Nh3\nB01\tScandinavian Defense\t1. e4 d5\n",
			wantCount: 2,
		},
		{
			name:      "header only",
			input:     "eco\tname\tpgn\n",
			wantCount: 0,
		},
		{
			name:    "malformed line",
			input:   "eco\tname\tpgn\nbadline\n",
			wantErr: true,
		},
		{
			name:      "empty input",
			input:     "",
			wantCount: 0,
		},
		{
			name:      "skips empty lines",
			input:     "eco\tname\tpgn\n\nA00\tAmar Opening\t1. Nh3\n\n",
			wantCount: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			entries, err := parseOpenings(strings.NewReader(tc.input))

			if tc.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, entries, tc.wantCount)
		})
	}
}

func TestParseLine(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		line     string
		wantName string
		wantECO  string
		wantErr  bool
	}{
		{
			name:     "simple opening",
			line:     "A00\tAmar Opening\t1. Nh3",
			wantName: "Amar Opening",
			wantECO:  "A00",
		},
		{
			name:     "multi-move opening",
			line:     "D01\tRapport-Jobava System\t1. d4 d5 2. Nc3 Nf6 3. Bf4",
			wantName: "Rapport-Jobava System",
			wantECO:  "D01",
		},
		{
			name:    "too few fields",
			line:    "A00\tAmar Opening",
			wantErr: true,
		},
		{
			name:    "invalid move",
			line:    "A00\tBad Opening\t1. Zz9",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			entry, err := parseLine(tc.line)

			if tc.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.wantECO, entry.opening.ECO)
			assert.Equal(t, tc.wantName, entry.opening.Name)
			assert.NotEmpty(t, entry.epd)
			assert.NotEmpty(t, entry.uci)
		})
	}
}

func TestParsePGNMoves(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		pgn  string
		want []string
	}{
		{
			name: "simple moves",
			pgn:  "1. e4 e5 2. Nf3 Nc6",
			want: []string{"e4", "e5", "Nf3", "Nc6"},
		},
		{
			name: "single move",
			pgn:  "1. d4",
			want: []string{"d4"},
		},
		{
			name: "with result",
			pgn:  "1. e4 e5 1-0",
			want: []string{"e4", "e5"},
		},
		{
			name: "compact format no space",
			pgn:  "1.e4 e5 2.Nf3 Nc6",
			want: []string{"e4", "e5", "Nf3", "Nc6"},
		},
		{
			name: "black continuation dots",
			pgn:  "1.e4 1...e5 2.Nf3",
			want: []string{"e4", "e5", "Nf3"},
		},
		{
			name: "compact with result",
			pgn:  "1.e4 e5 1-0",
			want: []string{"e4", "e5"},
		},
		{
			name: "empty",
			pgn:  "",
			want: []string{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := parsePGNMoves(tc.pgn)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestPositionToEPD(t *testing.T) {
	t.Parallel()

	entry, err := parseLine("A00\tTest\t1. e4")
	require.NoError(t, err)

	// After 1. e4 the EPD should have en passant on e3.
	assert.Contains(t, entry.epd, "e3")
	// Black to move.
	assert.Contains(t, entry.epd, " b ")
}

func TestStripPGNToMovetext(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		pgn  string
		want string
	}{
		{
			name: "plain movetext unchanged",
			pgn:  "1. e4 e5 2. Nf3 Nc6",
			want: "1. e4 e5 2. Nf3 Nc6",
		},
		{
			name: "strips tag pairs",
			pgn:  "[Event \"World Cup\"]\n[Site \"Sochi\"]\n\n1. e4 e5",
			want: "\n\n\n1. e4 e5",
		},
		{
			name: "strips brace comments",
			pgn:  "1. e4 {best move} e5 {solid} 2. Nf3",
			want: "1. e4  e5  2. Nf3",
		},
		{
			name: "strips semicolon comments",
			pgn:  "1. e4 e5 ; this is a comment\n2. Nf3",
			want: "1. e4 e5 2. Nf3",
		},
		{
			name: "strips NAGs",
			pgn:  "1. e4 $1 e5 $6 2. Nf3 $14",
			want: "1. e4  e5  2. Nf3 ",
		},
		{
			name: "strips variations",
			pgn:  "1. e4 e5 (1... c5 2. Nf3) 2. Nf3 Nc6",
			want: "1. e4 e5  2. Nf3 Nc6",
		},
		{
			name: "strips nested variations",
			pgn:  "1. e4 e5 (1... c5 (1... d5 2. exd5) 2. Nf3) 2. Nf3",
			want: "1. e4 e5  2. Nf3",
		},
		{
			name: "full PGN with everything",
			pgn:  "[Event \"Test\"]\n[Result \"1-0\"]\n\n1. e4 {!} e5 $1 (1... c5) 2. Nf3 ; knight\nNc6 1-0",
			want: "\n\n\n1. e4  e5   2. Nf3 Nc6 1-0",
		},
		{
			name: "empty input",
			pgn:  "",
			want: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := stripPGNToMovetext(tc.pgn)
			assert.Equal(t, tc.want, got)
		})
	}
}
