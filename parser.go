package openings

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/notnil/chess"
)

// openingEntry holds the parsed data for one opening, including computed fields.
type openingEntry struct {
	opening *Opening
	epd     string
	uci     []string
}

// parseOpenings reads a TSV file (eco\tname\tpgn) and returns parsed opening entries.
// It skips the header line and returns an error on the first malformed line.
func parseOpenings(r io.Reader) ([]*openingEntry, error) {
	scanner := bufio.NewScanner(r)

	// Skip header line.
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("reading header: %w", err)
		}

		return nil, nil
	}

	var entries []*openingEntry

	lineNum := 1

	for scanner.Scan() {
		lineNum++

		line := scanner.Text()
		if line == "" {
			continue
		}

		entry, err := parseLine(line)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNum, err)
		}

		entries = append(entries, entry)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading TSV: %w", err)
	}

	return entries, nil
}

// parseLine parses a single TSV line into an openingEntry.
// It replays the PGN moves on a chess board to compute EPD and UCI moves.
func parseLine(line string) (*openingEntry, error) {
	parts := strings.SplitN(line, "\t", 3)

	const expectedFields = 3
	if len(parts) != expectedFields {
		return nil, fmt.Errorf("expected %d tab-separated fields, got %d", expectedFields, len(parts))
	}

	eco := parts[0]
	name := parts[1]
	pgn := parts[2]

	epd, uciMoves, err := replayPGN(pgn)
	if err != nil {
		return nil, fmt.Errorf("replaying PGN for %s (%s): %w", name, pgn, err)
	}

	return &openingEntry{
		opening: &Opening{
			ECO:  eco,
			Name: name,
			PGN:  pgn,
		},
		epd: epd,
		uci: uciMoves,
	}, nil
}

// replayPGN replays PGN moves on a chess board and returns the final EPD
// and the sequence of UCI moves.
func replayPGN(pgn string) (epd string, uciMoves []string, err error) {
	moves := parsePGNMoves(pgn)
	if len(moves) == 0 {
		return "", nil, fmt.Errorf("no moves found in PGN: %q", pgn)
	}

	game := chess.NewGame()

	uciMoves = make([]string, 0, len(moves))

	for _, san := range moves {
		err = game.MoveStr(san)
		if err != nil {
			return "", nil, fmt.Errorf("invalid move %q: %w", san, err)
		}

		allMoves := game.Moves()
		lastMove := allMoves[len(allMoves)-1]
		uciMoves = append(uciMoves, lastMove.String())
	}

	epd = positionToEPD(game.Position())

	return epd, uciMoves, nil
}

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
