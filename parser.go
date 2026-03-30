package openings

import (
	"strings"

	"github.com/notnil/chess"

	"github.com/ksysoev/chess-openings/internal/epdutil"
)

// positionToEPD converts a chess position to EPD format (FEN without move counters).
// EPD consists of: piece placement, active color, castling rights, en passant square.
func positionToEPD(pos *chess.Position) string {
	return epdutil.PositionToEPD(pos)
}

// parsePGNMoves extracts individual SAN moves from a PGN move text string.
// It strips move numbers (e.g. "1.", "2.", "1.e4", "1...e5") and handles
// both spaced ("1. e4 e5") and compact ("1.e4 1...e5") PGN formats.
func parsePGNMoves(pgn string) []string {
	return epdutil.ParsePGNMoves(pgn)
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
