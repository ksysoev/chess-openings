// Package openings provides chess opening identification from move sequences.
//
// It uses the Lichess chess-openings database (~3,500 named openings) and
// identifies openings by matching board positions, which naturally handles
// transpositions. For example, reaching the Rapport-Jobava System via
// 1.d4 Nf6 2.Nc3 d5 3.Bf4 (transposed) is correctly identified the same
// as the standard 1.d4 d5 2.Nc3 Nf6 3.Bf4.
package openings

// Opening represents a named chess opening position.
type Opening struct {
	// ECO is the Encyclopedia of Chess Openings classification code (e.g. "D01").
	ECO string
	// Name is the full opening name (e.g. "Rapport-Jobava System").
	Name string
	// PGN is the canonical move sequence in standard algebraic notation.
	PGN string
}

// Classification is the result of identifying a game's opening.
type Classification struct {
	// Opening is the identified opening. Nil if no known opening was found.
	Opening *Opening
	// Ply is the half-move depth at which the opening was identified.
	// For example, after 1.e4 e5 the ply is 2.
	Ply int
}
