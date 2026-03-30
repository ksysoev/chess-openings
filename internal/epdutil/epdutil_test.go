package epdutil_test

import (
	"strings"
	"testing"

	"github.com/notnil/chess"
	"github.com/stretchr/testify/assert"

	"github.com/ksysoev/chess-openings/internal/epdutil"
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

			got := epdutil.ParsePGNMoves(tc.pgn)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestPositionToEPD(t *testing.T) {
	t.Parallel()

	game := chess.NewGame()
	err := game.MoveStr("e4")

	assert.NoError(t, err)

	epd := epdutil.PositionToEPD(game.Position())

	// After 1. e4 the EPD should have en passant on e3.
	assert.Contains(t, epd, "e3")
	// Black to move.
	assert.Contains(t, epd, " b ")
}

func TestPositionToEPD_StartingPosition(t *testing.T) {
	t.Parallel()

	game := chess.NewGame()
	epd := epdutil.PositionToEPD(game.Position())

	// Starting position: white to move, full castling, no en passant.
	assert.Contains(t, epd, " w ")
	assert.Contains(t, epd, "KQkq")
	assert.True(t, strings.HasSuffix(epd, " -"), "EPD should end with no en passant marker")
}
