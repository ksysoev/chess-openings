# chess-openings

[![Tests](https://github.com/ksysoev/chess-openings/actions/workflows/tests.yml/badge.svg)](https://github.com/ksysoev/chess-openings/actions/workflows/tests.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/ksysoev/chess-openings)](https://goreportcard.com/report/github.com/ksysoev/chess-openings)
[![Go Reference](https://pkg.go.dev/badge/github.com/ksysoev/chess-openings.svg)](https://pkg.go.dev/github.com/ksysoev/chess-openings)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

A Go library for identifying chess openings from move sequences. It uses the [Lichess chess-openings database](https://github.com/lichess-org/chess-openings) (~3,500 named openings across all ECO codes) and identifies openings by matching board positions, which naturally handles transpositions.

## Features

- Position-based matching that handles transpositions (same position via different move orders)
- Accepts UCI, SAN, PGN, FEN, and EPD input formats
- Embedded database with no external files or network access required at runtime
- ~3,500 named openings covering ECO codes A through E
- Includes modern openings

## Installation

```sh
go get github.com/ksysoev/chess-openings@latest
```

## Usage

```go
package main

import (
	"fmt"
	"log"

	openings "github.com/ksysoev/chess-openings"
)

func main() {
	book := openings.New()

	// Classify from SAN moves
	result, err := book.ClassifySAN([]string{"e4", "c5", "Nf3", "d6"})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s %s\n", result.Opening.ECO, result.Opening.Name)
	// Output: B50 Sicilian Defense: Modern Variations

	// Classify from UCI moves
	result, err = book.Classify([]string{"d2d4", "d7d5", "b1c3", "g8f6", "c1f4"})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s %s\n", result.Opening.ECO, result.Opening.Name)
	// Output: D01 Rapport-Jobava System

	// Classify from a PGN string
	result, err = book.ClassifyPGN("1. e4 e5 2. Nf3 Nc6 3. Bb5")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s %s\n", result.Opening.ECO, result.Opening.Name)
	// Output: C60 Ruy Lopez

	// Look up a position by FEN
	opening, found := book.ClassifyPosition(
		"rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1",
	)
	if found {
		fmt.Printf("%s %s\n", opening.ECO, opening.Name)
	}
}
```

## API

| Method | Input | Description |
|---|---|---|
| `New()` | - | Creates a `Book` loaded with the full Lichess database |
| `Classify(uciMoves)` | UCI move strings | Identifies opening with transposition support |
| `ClassifySAN(sanMoves)` | SAN move strings | Same as `Classify` but with SAN input |
| `ClassifyPGN(pgn)` | PGN string | Parses PGN (with optional tags/comments) and identifies the opening |
| `ClassifyPosition(fen)` | FEN string | Looks up the opening for a board position |
| `LookupPosition(epd)` | EPD string | Direct position lookup in the database |
| `LookupMoves(uciMoves)` | UCI move strings | Exact move sequence lookup (no transpositions) |
| `SearchMoves(uciMoves)` | UCI move strings | Deepest match along a move sequence (no transpositions) |
| `Size()` | - | Returns the number of unique positions in the book |

## Data Source

Opening data is from the [Lichess chess-openings](https://github.com/lichess-org/chess-openings) project, licensed under [CC0 1.0](https://creativecommons.org/publicdomain/zero/1.0/). The data files are embedded in the binary at compile time.

## License

chess-openings is licensed under the MIT License. See the LICENSE file for more details.
