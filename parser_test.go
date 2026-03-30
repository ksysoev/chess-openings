package openings

import (
	"testing"

	"github.com/notnil/chess"
	"github.com/stretchr/testify/assert"
)

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

	game := chess.NewGame()
	err := game.MoveStr("e4")

	assert.NoError(t, err)

	epd := positionToEPD(game.Position())

	// After 1. e4 the EPD should have en passant on e3.
	assert.Contains(t, epd, "e3")
	// Black to move.
	assert.Contains(t, epd, " b ")
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
