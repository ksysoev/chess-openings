// Package epdutil provides shared chess position and PGN parsing utilities
// used by both the openings library and the code generator (cmd/generate).
package epdutil

import (
	"strings"

	"github.com/notnil/chess"
)

// PositionToEPD converts a chess position to EPD format (FEN without move counters).
// EPD consists of: piece placement, active color, castling rights, en passant square.
func PositionToEPD(pos *chess.Position) string {
	board := pos.Board().String()
	turn := pos.Turn().String()
	castle := pos.CastleRights().String()

	ep := "-"
	if pos.EnPassantSquare() != chess.NoSquare {
		ep = pos.EnPassantSquare().String()
	}

	return board + " " + turn + " " + castle + " " + ep
}

// ParsePGNMoves extracts individual SAN moves from a PGN move text string.
// It strips move numbers (e.g. "1.", "2.", "1.e4", "1...e5") and handles
// both spaced ("1. e4 e5") and compact ("1.e4 1...e5") PGN formats.
func ParsePGNMoves(pgn string) []string {
	tokens := strings.Fields(pgn)
	moves := make([]string, 0, len(tokens))

	for _, token := range tokens {
		// Skip result markers.
		if token == "1-0" || token == "0-1" || token == "1/2-1/2" || token == "*" {
			continue
		}

		// Strip move-number prefix: "1.", "12.", "1...", "1.e4", "1...e5"
		// Find the last dot and take everything after it.
		if dotIdx := strings.LastIndex(token, "."); dotIdx >= 0 {
			token = token[dotIdx+1:]
		}

		if token != "" {
			moves = append(moves, token)
		}
	}

	return moves
}
