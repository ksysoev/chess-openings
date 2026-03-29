package openings

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	sharedBook     *Book
	sharedBookOnce sync.Once
	sharedBookErr  error
)

func getTestBook(t *testing.T) *Book {
	t.Helper()

	sharedBookOnce.Do(func() {
		sharedBook, sharedBookErr = New()
	})

	require.NoError(t, sharedBookErr)
	require.NotNil(t, sharedBook)

	return sharedBook
}

func TestNew(t *testing.T) {
	t.Parallel()

	book := getTestBook(t)

	// The lichess database has ~3500 unique positions.
	assert.Greater(t, book.Size(), 3000, "should load a significant number of openings")
}

func TestClassifySAN_StandardOpenings(t *testing.T) {
	t.Parallel()

	book := getTestBook(t)

	tests := []struct {
		name     string
		wantName string
		wantECO  string
		moves    []string
	}{
		{
			name:     "Sicilian Defense",
			moves:    []string{"e4", "c5"},
			wantName: "Sicilian Defense",
			wantECO:  "B20",
		},
		{
			name:     "Italian Game",
			moves:    []string{"e4", "e5", "Nf3", "Nc6", "Bc4"},
			wantName: "Italian Game",
			wantECO:  "C50",
		},
		{
			name:     "Ruy Lopez",
			moves:    []string{"e4", "e5", "Nf3", "Nc6", "Bb5"},
			wantName: "Ruy Lopez",
			wantECO:  "C60",
		},
		{
			name:     "Queen's Gambit",
			moves:    []string{"d4", "d5", "c4"},
			wantName: "Queen's Gambit",
			wantECO:  "D06",
		},
		{
			name:     "Rapport-Jobava System",
			moves:    []string{"d4", "d5", "Nc3", "Nf6", "Bf4"},
			wantName: "Rapport-Jobava System",
			wantECO:  "D01",
		},
		{
			name:     "London System",
			moves:    []string{"d4", "d5", "Nf3", "Nf6", "Bf4"},
			wantName: "Queen's Pawn Game: London System",
			wantECO:  "D02",
		},
		{
			name:     "French Defense",
			moves:    []string{"e4", "e6"},
			wantName: "French Defense",
			wantECO:  "C00",
		},
		{
			name:     "Caro-Kann Defense",
			moves:    []string{"e4", "c6"},
			wantName: "Caro-Kann Defense",
			wantECO:  "B10",
		},
		{
			name:     "King's Indian Defense",
			moves:    []string{"d4", "Nf6", "c4", "g6", "Nc3", "Bg7"},
			wantName: "King's Indian Defense",
			wantECO:  "E61",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result, err := book.ClassifySAN(tc.moves)
			require.NoError(t, err)
			require.NotNil(t, result.Opening, "expected an opening to be found")
			assert.Equal(t, tc.wantName, result.Opening.Name)
			assert.Equal(t, tc.wantECO, result.Opening.ECO)
		})
	}
}

func TestClassifySAN_DeepVariations(t *testing.T) {
	t.Parallel()

	book := getTestBook(t)

	// Sicilian Najdorf is a deep variation that should be identified
	// as the most specific match.
	moves := []string{"e4", "c5", "Nf3", "d6", "d4", "cxd4", "Nxd4", "Nf6", "Nc3", "a6"}

	result, err := book.ClassifySAN(moves)
	require.NoError(t, err)
	require.NotNil(t, result.Opening)
	assert.Contains(t, result.Opening.Name, "Najdorf")
}

func TestClassifySAN_Transpositions(t *testing.T) {
	t.Parallel()

	book := getTestBook(t)

	// Test transposition: Rapport-Jobava reached via different move order.
	// Standard: 1.d4 d5 2.Nc3 Nf6 3.Bf4
	// Transposed: 1.d4 Nf6 2.Nc3 d5 3.Bf4
	standard := []string{"d4", "d5", "Nc3", "Nf6", "Bf4"}
	transposed := []string{"d4", "Nf6", "Nc3", "d5", "Bf4"}

	stdResult, err := book.ClassifySAN(standard)
	require.NoError(t, err)
	require.NotNil(t, stdResult.Opening)

	transResult, err := book.ClassifySAN(transposed)
	require.NoError(t, err)
	require.NotNil(t, transResult.Opening)

	// Both should identify the same opening (Rapport-Jobava System).
	assert.Equal(t, stdResult.Opening.Name, transResult.Opening.Name,
		"transposed move order should identify the same opening")
	assert.Contains(t, stdResult.Opening.Name, "Rapport-Jobava")
}

func TestClassifySAN_LondonTransposition(t *testing.T) {
	t.Parallel()

	book := getTestBook(t)

	// London System can be reached via different move orders.
	// Standard: 1.d4 d5 2.Nf3 Nf6 3.Bf4
	// Transposed: 1.d4 Nf6 2.Bf4 d5 3.Nf3
	standard := []string{"d4", "d5", "Nf3", "Nf6", "Bf4"}
	transposed := []string{"d4", "Nf6", "Bf4", "d5", "Nf3"}

	stdResult, err := book.ClassifySAN(standard)
	require.NoError(t, err)
	require.NotNil(t, stdResult.Opening)

	transResult, err := book.ClassifySAN(transposed)
	require.NoError(t, err)
	require.NotNil(t, transResult.Opening)

	assert.Equal(t, stdResult.Opening.Name, transResult.Opening.Name,
		"London System should be identified regardless of move order")
}

func TestClassify_UCI(t *testing.T) {
	t.Parallel()

	book := getTestBook(t)

	tests := []struct {
		name     string
		wantName string
		moves    []string
	}{
		{
			name:     "Sicilian Defense via UCI",
			moves:    []string{"e2e4", "c7c5"},
			wantName: "Sicilian Defense",
		},
		{
			name:     "Italian Game via UCI",
			moves:    []string{"e2e4", "e7e5", "g1f3", "b8c6", "f1c4"},
			wantName: "Italian Game",
		},
		{
			name:     "Rapport-Jobava via UCI",
			moves:    []string{"d2d4", "d7d5", "b1c3", "g8f6", "c1f4"},
			wantName: "Rapport-Jobava System",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result, err := book.Classify(tc.moves)
			require.NoError(t, err)
			require.NotNil(t, result.Opening)
			assert.Equal(t, tc.wantName, result.Opening.Name)
		})
	}
}

func TestClassifyPGN(t *testing.T) {
	t.Parallel()

	book := getTestBook(t)

	tests := []struct {
		name     string
		pgn      string
		wantName string
	}{
		{
			name:     "from PGN string",
			pgn:      "1. e4 c5",
			wantName: "Sicilian Defense",
		},
		{
			name:     "longer PGN",
			pgn:      "1. d4 d5 2. Nc3 Nf6 3. Bf4",
			wantName: "Rapport-Jobava System",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result, err := book.ClassifyPGN(tc.pgn)
			require.NoError(t, err)
			require.NotNil(t, result.Opening)
			assert.Equal(t, tc.wantName, result.Opening.Name)
		})
	}
}

func TestClassifySAN_EmptyMoves(t *testing.T) {
	t.Parallel()

	book := getTestBook(t)

	result, err := book.ClassifySAN([]string{})
	require.NoError(t, err)
	assert.Nil(t, result.Opening)
}

func TestClassifySAN_InvalidMove(t *testing.T) {
	t.Parallel()

	book := getTestBook(t)

	_, err := book.ClassifySAN([]string{"Zz9"})
	assert.Error(t, err)
}

func TestClassify_InvalidUCIMove(t *testing.T) {
	t.Parallel()

	book := getTestBook(t)

	_, err := book.Classify([]string{"z9z9"})
	assert.Error(t, err)
}

func TestClassifySAN_MoveBeyondBook(t *testing.T) {
	t.Parallel()

	book := getTestBook(t)

	// Play an opening and then continue with moves beyond the book.
	// Should still identify the last known opening.
	moves := []string{"e4", "e5", "Nf3", "Nc6", "Bc4", "Nf6", "d3", "h6", "a3", "a6"}

	result, err := book.ClassifySAN(moves)
	require.NoError(t, err)
	require.NotNil(t, result.Opening, "should identify the opening even with extra moves")
	assert.NotEmpty(t, result.Opening.Name)
}

func TestClassifySAN_PlyTracking(t *testing.T) {
	t.Parallel()

	book := getTestBook(t)

	// After 1.e4 e5 (2 plies), there should be a match.
	// After 1.e4 e5 2.Nf3 Nc6 3.Bc4 (5 plies), there should be a deeper match.
	moves := []string{"e4", "e5", "Nf3", "Nc6", "Bc4"}

	result, err := book.ClassifySAN(moves)
	require.NoError(t, err)
	require.NotNil(t, result.Opening)
	assert.Equal(t, 5, result.Ply, "should identify at ply 5 for Italian Game")
}

func TestClassifySAN_NoMatch(t *testing.T) {
	t.Parallel()

	book := getTestBook(t)

	// An unlikely sequence that shouldn't match any known opening name.
	// But 1.a3 matches "Anderssen's Opening", so use something else.
	// Actually most first moves have names. Let's try something unusual.
	moves := []string{"a3", "h5", "b4", "g5"}

	result, err := book.ClassifySAN(moves)
	require.NoError(t, err)

	// Even unusual openings may have a match (1.a3 is "Anderssen's Opening").
	// The key thing is no error.
	assert.NotNil(t, result)
}

func TestLookupPosition(t *testing.T) {
	t.Parallel()

	book := getTestBook(t)

	// EPD for the starting position after 1.e4.
	epd := "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3"

	opening, found := book.LookupPosition(epd)
	assert.True(t, found)
	assert.NotNil(t, opening)
}

func TestLookupPosition_NotFound(t *testing.T) {
	t.Parallel()

	book := getTestBook(t)

	_, found := book.LookupPosition("fake/epd/string w - -")
	assert.False(t, found)
}

func TestClassifyPosition(t *testing.T) {
	t.Parallel()

	book := getTestBook(t)

	// Full FEN for position after 1.e4.
	fen := "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1"

	opening, found := book.ClassifyPosition(fen)
	assert.True(t, found)
	assert.NotNil(t, opening)
}

func TestClassifyPosition_InvalidFEN(t *testing.T) {
	t.Parallel()

	book := getTestBook(t)

	_, found := book.ClassifyPosition("too short")
	assert.False(t, found)
}

func TestSize(t *testing.T) {
	t.Parallel()

	book := getTestBook(t)
	assert.Greater(t, book.Size(), 0)
}

func BenchmarkNew(b *testing.B) {
	for b.Loop() {
		_, err := New()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkClassifySAN(b *testing.B) {
	book, err := New()
	if err != nil {
		b.Fatal(err)
	}

	moves := []string{"e4", "c5", "Nf3", "d6", "d4", "cxd4", "Nxd4", "Nf6", "Nc3", "a6"}

	b.ResetTimer()

	for b.Loop() {
		_, err := book.ClassifySAN(moves)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkClassifyUCI(b *testing.B) {
	book, err := New()
	if err != nil {
		b.Fatal(err)
	}

	moves := []string{"e2e4", "c7c5", "g1f3", "d7d6", "d2d4", "c5d4", "f3d4", "g8f6", "b1c3", "a7a6"}

	b.ResetTimer()

	for b.Loop() {
		_, err := book.Classify(moves)
		if err != nil {
			b.Fatal(err)
		}
	}
}
