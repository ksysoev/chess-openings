package openings

import (
	"strings"

	"github.com/notnil/chess"
)

// positionToEPD converts a chess position to EPD format (FEN without move counters).
// EPD consists of: piece placement, active color, castling rights, en passant square.
func positionToEPD(pos *chess.Position) string {
	board := pos.Board().String()
	turn := pos.Turn().String()
	castle := pos.CastleRights().String()

	ep := "-"
	if pos.EnPassantSquare() != chess.NoSquare {
		ep = pos.EnPassantSquare().String()
	}

	return board + " " + turn + " " + castle + " " + ep
}

// parsePGNMoves extracts individual SAN moves from a PGN move text string.
// It strips move numbers (e.g. "1.", "2.", "1.e4", "1...e5") and handles
// both spaced ("1. e4 e5") and compact ("1.e4 1...e5") PGN formats.
func parsePGNMoves(pgn string) []string {
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

// stripPGNToMovetext removes PGN tag pairs, comments, NAGs, and variations,
// returning only the movetext portion suitable for parsePGNMoves.
func stripPGNToMovetext(pgn string) string {
	var buf strings.Builder

	buf.Grow(len(pgn))

	braceDepth := 0 // {comment} nesting
	parenDepth := 0 // (variation) nesting
	inTag := false  // [Tag "value"] line

	for i := 0; i < len(pgn); i++ {
		ch := pgn[i]

		switch {
		case ch == '[' && braceDepth == 0 && parenDepth == 0:
			inTag = true
		case inTag:
			if ch == ']' {
				inTag = false
			}
		case ch == '{':
			braceDepth++
		case braceDepth > 0:
			if ch == '}' {
				braceDepth--
			}
		case ch == '(':
			parenDepth++
		case parenDepth > 0:
			if ch == ')' {
				parenDepth--
			}
		case ch == '$':
			// Skip NAG: $N where N is one or more digits.
			i++

			for i < len(pgn) && pgn[i] >= '0' && pgn[i] <= '9' {
				i++
			}

			i-- // loop will increment
		case ch == ';':
			// Skip rest-of-line comment.
			for i < len(pgn) && pgn[i] != '\n' {
				i++
			}
		default:
			buf.WriteByte(ch)
		}
	}

	return buf.String()
}
