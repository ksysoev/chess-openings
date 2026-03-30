// Command generate fetches the Lichess chess-openings database and produces
// a Go source file with pre-computed EPD positions and UCI move sequences.
// This eliminates runtime PGN parsing and board replay from the library's
// initialisation path.
//
// Usage:
//
//	go run ./cmd/generate
package main

import (
	"bufio"
	"fmt"
	"go/format"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/notnil/chess"
)

const (
	baseURL    = "https://raw.githubusercontent.com/lichess-org/chess-openings/master"
	outputFile = "openings_gen.go"
)

var tsvFiles = []string{"a", "b", "c", "d", "e"}

// entry holds the pre-computed data for a single opening.
type entry struct {
	eco  string
	name string
	pgn  string
	epd  string
	uci  []string
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("generate: ")

	var all []entry

	for _, name := range tsvFiles {
		url := fmt.Sprintf("%s/%s.tsv", baseURL, name)

		log.Printf("fetching %s.tsv ...", name)

		entries, err := fetchAndParse(url)
		if err != nil {
			log.Fatalf("processing %s.tsv: %v", name, err)
		}

		all = append(all, entries...)
		log.Printf("  %d openings from %s.tsv", len(entries), name)
	}

	log.Printf("total: %d openings", len(all))

	src := generate(all)

	formatted, err := format.Source(src)
	if err != nil {
		// Write the unformatted source for debugging.
		_ = os.WriteFile(outputFile, src, 0o600)
		log.Fatalf("gofmt failed (unformatted source written to %s): %v", outputFile, err)
	}

	if err := os.WriteFile(outputFile, formatted, 0o600); err != nil {
		log.Fatalf("writing %s: %v", outputFile, err)
	}

	log.Printf("wrote %s (%d bytes)", outputFile, len(formatted))
}

// fetchAndParse downloads a TSV file and parses every line into entries.
func fetchAndParse(url string) ([]entry, error) {
	resp, err := http.Get(url) //nolint:gosec,noctx // static trusted URL, no context needed
	if err != nil {
		return nil, fmt.Errorf("fetching %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching %s: HTTP %d", url, resp.StatusCode)
	}

	return parseTSV(resp.Body)
}

// parseTSV reads a TSV stream (eco\tname\tpgn) and returns parsed entries.
func parseTSV(r io.Reader) ([]entry, error) {
	scanner := bufio.NewScanner(r)

	// Skip header line.
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("reading header: %w", err)
		}

		return nil, nil
	}

	var entries []entry

	lineNum := 1

	for scanner.Scan() {
		lineNum++

		line := scanner.Text()
		if line == "" {
			continue
		}

		e, err := parseLine(line)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNum, err)
		}

		entries = append(entries, e)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading TSV: %w", err)
	}

	return entries, nil
}

// parseLine parses a single TSV line and replays the PGN to compute EPD and UCI.
func parseLine(line string) (entry, error) {
	const expectedFields = 3

	parts := strings.SplitN(line, "\t", expectedFields)
	if len(parts) != expectedFields {
		return entry{}, fmt.Errorf("expected %d tab-separated fields, got %d", expectedFields, len(parts))
	}

	eco := parts[0]
	name := parts[1]
	pgn := parts[2]

	epd, uciMoves, err := replayPGN(pgn)
	if err != nil {
		return entry{}, fmt.Errorf("replaying PGN for %s (%s): %w", name, pgn, err)
	}

	return entry{
		eco:  eco,
		name: name,
		pgn:  pgn,
		epd:  epd,
		uci:  uciMoves,
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
		if err = game.MoveStr(san); err != nil {
			return "", nil, fmt.Errorf("invalid move %q: %w", san, err)
		}

		allMoves := game.Moves()
		lastMove := allMoves[len(allMoves)-1]
		uciMoves = append(uciMoves, lastMove.String())
	}

	epd = positionToEPD(game.Position())

	return epd, uciMoves, nil
}

// parsePGNMoves extracts individual SAN moves from a PGN movetext string.
func parsePGNMoves(pgn string) []string {
	tokens := strings.Fields(pgn)
	moves := make([]string, 0, len(tokens))

	for _, token := range tokens {
		if token == "1-0" || token == "0-1" || token == "1/2-1/2" || token == "*" {
			continue
		}

		if dotIdx := strings.LastIndex(token, "."); dotIdx >= 0 {
			token = token[dotIdx+1:]
		}

		if token != "" {
			moves = append(moves, token)
		}
	}

	return moves
}

// positionToEPD converts a chess position to EPD format (FEN without move counters).
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

// generate produces the Go source for the pre-computed openings file.
func generate(entries []entry) []byte {
	var buf strings.Builder

	buf.WriteString("// Code generated by cmd/generate; DO NOT EDIT.\n\n")
	buf.WriteString("package openings\n\n")

	// Type definition for the generated entries.
	buf.WriteString("// generatedEntry holds pre-computed data for a single chess opening.\n")
	buf.WriteString("// EPD and UCI moves are computed at generation time by replaying the PGN,\n")
	buf.WriteString("// eliminating the need for runtime board replay.\n")
	buf.WriteString("type generatedEntry struct {\n")
	buf.WriteString("\teco  string\n")
	buf.WriteString("\tname string\n")
	buf.WriteString("\tpgn  string\n")
	buf.WriteString("\tepd  string\n")
	buf.WriteString("\tuci  []string\n")
	buf.WriteString("}\n\n")

	// Opening data as a fixed-size array.
	fmt.Fprintf(&buf, "// generatedOpenings contains %d pre-computed chess openings from the\n", len(entries))
	buf.WriteString("// Lichess opening database (https://github.com/lichess-org/chess-openings).\n")
	fmt.Fprintf(&buf, "var generatedOpenings = [%d]generatedEntry{\n", len(entries))

	for _, e := range entries {
		fmt.Fprintf(&buf, "\t{%q, %q, %q, %q, []string{", e.eco, e.name, e.pgn, e.epd)

		for i, m := range e.uci {
			if i > 0 {
				buf.WriteString(", ")
			}

			fmt.Fprintf(&buf, "%q", m)
		}

		buf.WriteString("}},\n")
	}

	buf.WriteString("}\n")

	return []byte(buf.String())
}
