package openings

import (
	"embed"
	"fmt"
	"io/fs"
	"strings"

	"github.com/notnil/chess"
)

//go:embed data/*.tsv
var dataFS embed.FS

// Book is a chess opening identification engine loaded with the Lichess
// opening database. It identifies openings by matching board positions,
// which naturally handles transpositions.
type Book struct {
	positions map[string]*Opening // EPD -> Opening (position-based lookup)
	trie      *trieNode           // UCI move trie (sequence-based lookup)
}

// New creates a new Book loaded with the full Lichess opening database.
// The database contains ~3,500 named openings across all ECO codes (A-E).
func New() (*Book, error) {
	book := &Book{
		positions: make(map[string]*Opening),
		trie:      newTrieNode(),
	}

	files, err := fs.Glob(dataFS, "data/*.tsv")
	if err != nil {
		return nil, fmt.Errorf("listing data files: %w", err)
	}

	for _, path := range files {
		if err := book.loadFile(path); err != nil {
			return nil, fmt.Errorf("loading %s: %w", path, err)
		}
	}

	return book, nil
}

// loadFile parses a single TSV file and adds its entries to the book.
func (b *Book) loadFile(path string) error {
	file, err := dataFS.Open(path)
	if err != nil {
		return fmt.Errorf("opening file: %w", err)
	}
	defer file.Close()

	entries, err := parseOpenings(file)
	if err != nil {
		return fmt.Errorf("parsing: %w", err)
	}

	for _, entry := range entries {
		b.addEntry(entry)
	}

	return nil
}

// addEntry adds a single opening entry to both the position map and the trie.
// For the position map, if an EPD already exists, the entry with more moves
// (more specific opening) is kept to provide the most precise classification.
func (b *Book) addEntry(entry *openingEntry) {
	existing, exists := b.positions[entry.epd]
	if !exists || len(entry.uci) > len(parsePGNMoves(existing.PGN)) {
		b.positions[entry.epd] = entry.opening
	}

	b.trie.insert(entry.uci, entry.opening)
}

// Classify identifies the opening from a sequence of moves in UCI notation
// (e.g. "e2e4", "d7d5", "c2c4"). It replays the moves on a chess board and
// checks each resulting position against the opening database.
//
// Returns the most specific (deepest) opening found. This approach naturally
// handles transpositions: if a game reaches a known opening position via a
// non-standard move order, it will still be correctly identified.
//
// Returns a Classification with a nil Opening if no known opening was found.
func (b *Book) Classify(uciMoves []string) (*Classification, error) {
	if len(uciMoves) == 0 {
		return &Classification{}, nil
	}

	game := chess.NewGame(chess.UseNotation(chess.UCINotation{}))
	uci := chess.UCINotation{}

	var best *Classification

	for i, moveStr := range uciMoves {
		move, err := uci.Decode(game.Position(), moveStr)
		if err != nil {
			return nil, fmt.Errorf("invalid UCI move %q at ply %d: %w", moveStr, i+1, err)
		}

		if err := game.Move(move); err != nil {
			return nil, fmt.Errorf("illegal move %q at ply %d: %w", moveStr, i+1, err)
		}

		epd := positionToEPD(game.Position())
		if opening, found := b.positions[epd]; found {
			best = &Classification{
				Opening: opening,
				Ply:     i + 1,
			}
		}
	}

	if best == nil {
		return &Classification{}, nil
	}

	return best, nil
}

// ClassifySAN identifies the opening from a sequence of moves in Standard
// Algebraic Notation (e.g. "e4", "d5", "c4"). It works the same as Classify
// but accepts SAN input.
func (b *Book) ClassifySAN(sanMoves []string) (*Classification, error) {
	if len(sanMoves) == 0 {
		return &Classification{}, nil
	}

	game := chess.NewGame()

	var best *Classification

	for i, san := range sanMoves {
		if err := game.MoveStr(san); err != nil {
			return nil, fmt.Errorf("invalid SAN move %q at ply %d: %w", san, i+1, err)
		}

		epd := positionToEPD(game.Position())
		if opening, found := b.positions[epd]; found {
			best = &Classification{
				Opening: opening,
				Ply:     i + 1,
			}
		}
	}

	if best == nil {
		return &Classification{}, nil
	}

	return best, nil
}

// ClassifyPGN identifies the opening from a PGN move text string
// (e.g. "1. e4 e5 2. Nf3 Nc6"). Move numbers are stripped automatically.
func (b *Book) ClassifyPGN(pgn string) (*Classification, error) {
	moves := parsePGNMoves(pgn)

	return b.ClassifySAN(moves)
}

// LookupPosition finds the opening for a given EPD position string.
// EPD format is FEN without the halfmove clock and fullmove number fields:
// "<piece-placement> <active-color> <castling> <en-passant>".
func (b *Book) LookupPosition(epd string) (*Opening, bool) {
	opening, found := b.positions[epd]

	return opening, found
}

// LookupMoves finds the opening matching the exact UCI move sequence in the trie.
// Unlike Classify, this does not check positions and does not handle transpositions.
func (b *Book) LookupMoves(uciMoves []string) (*Opening, bool) {
	return b.trie.lookup(uciMoves)
}

// Size returns the number of unique positions in the opening book.
func (b *Book) Size() int {
	return len(b.positions)
}

// SearchMoves walks the trie following UCI moves and returns the deepest
// opening found along the path. Unlike Classify, this only matches by
// exact move sequence without position-based transposition detection.
func (b *Book) SearchMoves(uciMoves []string) *Opening {
	return b.trie.search(uciMoves)
}

// ClassifyPosition computes the EPD for the given FEN string and looks it up.
// This is a convenience method for users who have a FEN string instead of EPD.
func (b *Book) ClassifyPosition(fen string) (*Opening, bool) {
	// EPD is the first 4 fields of FEN.
	parts := strings.SplitN(fen, " ", 6)

	const minEPDFields = 4
	if len(parts) < minEPDFields {
		return nil, false
	}

	epd := strings.Join(parts[:minEPDFields], " ")

	return b.LookupPosition(epd)
}
