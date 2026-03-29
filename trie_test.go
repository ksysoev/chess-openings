package openings

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrieInsertAndLookup(t *testing.T) {
	t.Parallel()

	root := newTrieNode()

	opening := &Opening{ECO: "C50", Name: "Italian Game", PGN: "1. e4 e5 2. Nf3 Nc6 3. Bc4"}
	root.insert([]string{"e2e4", "e7e5", "g1f3", "b8c6", "f1c4"}, opening)

	tests := []struct {
		want  *Opening
		name  string
		moves []string
		found bool
	}{
		{
			name:  "exact match",
			moves: []string{"e2e4", "e7e5", "g1f3", "b8c6", "f1c4"},
			want:  opening,
			found: true,
		},
		{
			name:  "partial match returns nil",
			moves: []string{"e2e4", "e7e5"},
			want:  nil,
			found: false,
		},
		{
			name:  "no match",
			moves: []string{"d2d4"},
			want:  nil,
			found: false,
		},
		{
			name:  "empty moves",
			moves: []string{},
			want:  nil,
			found: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, found := root.lookup(tc.moves)
			assert.Equal(t, tc.found, found)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestTrieSearch(t *testing.T) {
	t.Parallel()

	root := newTrieNode()

	kp := &Opening{ECO: "B00", Name: "King's Pawn Game", PGN: "1. e4 e5"}
	italian := &Opening{ECO: "C50", Name: "Italian Game", PGN: "1. e4 e5 2. Nf3 Nc6 3. Bc4"}

	root.insert([]string{"e2e4", "e7e5"}, kp)
	root.insert([]string{"e2e4", "e7e5", "g1f3", "b8c6", "f1c4"}, italian)

	tests := []struct {
		want  *Opening
		name  string
		moves []string
	}{
		{
			name:  "matches deepest opening",
			moves: []string{"e2e4", "e7e5", "g1f3", "b8c6", "f1c4"},
			want:  italian,
		},
		{
			name:  "matches intermediate opening",
			moves: []string{"e2e4", "e7e5", "g1f3"},
			want:  kp,
		},
		{
			name:  "matches early opening",
			moves: []string{"e2e4", "e7e5"},
			want:  kp,
		},
		{
			name:  "continues beyond known moves",
			moves: []string{"e2e4", "e7e5", "g1f3", "b8c6", "f1c4", "g8f6"},
			want:  italian,
		},
		{
			name:  "no match at all",
			moves: []string{"d2d4"},
			want:  nil,
		},
		{
			name:  "empty moves",
			moves: []string{},
			want:  nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := root.search(tc.moves)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestTrieMultipleOpenings(t *testing.T) {
	t.Parallel()

	root := newTrieNode()

	sicilian := &Opening{ECO: "B20", Name: "Sicilian Defense", PGN: "1. e4 c5"}
	najdorf := &Opening{ECO: "B90", Name: "Sicilian Defense: Najdorf Variation", PGN: "1. e4 c5 2. Nf3 d6 3. d4 cxd4 4. Nxd4 Nf6 5. Nc3 a6"}

	root.insert([]string{"e2e4", "c7c5"}, sicilian)
	root.insert([]string{"e2e4", "c7c5", "g1f3", "d7d6", "d2d4", "c5d4", "f3d4", "g8f6", "b1c3", "a7a6"}, najdorf)

	// Searching with full Najdorf moves should find Najdorf.
	got := root.search([]string{"e2e4", "c7c5", "g1f3", "d7d6", "d2d4", "c5d4", "f3d4", "g8f6", "b1c3", "a7a6"})
	assert.Equal(t, najdorf, got)

	// Searching with just 1.e4 c5 should find Sicilian.
	got = root.search([]string{"e2e4", "c7c5"})
	assert.Equal(t, sicilian, got)
}
