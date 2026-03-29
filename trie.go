package openings

// trieNode is a node in a trie indexed by UCI move strings.
// Each node may optionally hold an Opening if the move sequence
// ending at this node is a named opening.
type trieNode struct {
	children map[string]*trieNode
	opening  *Opening
}

// newTrieNode creates a new empty trie node.
func newTrieNode() *trieNode {
	return &trieNode{
		children: make(map[string]*trieNode),
	}
}

// insert adds an opening at the given UCI move path in the trie.
func (n *trieNode) insert(uciMoves []string, opening *Opening) {
	current := n

	for _, move := range uciMoves {
		child, exists := current.children[move]
		if !exists {
			child = newTrieNode()
			current.children[move] = child
		}

		current = child
	}

	current.opening = opening
}

// search walks the trie following the given UCI moves and returns
// the deepest opening found along the path. Returns nil if no
// opening is found at any point along the move sequence.
func (n *trieNode) search(uciMoves []string) *Opening {
	current := n

	var best *Opening

	for _, move := range uciMoves {
		child, exists := current.children[move]
		if !exists {
			break
		}

		current = child

		if current.opening != nil {
			best = current.opening
		}
	}

	return best
}

// lookup walks the trie following the exact UCI move sequence and returns
// the opening at that exact node. Returns nil and false if the exact
// sequence is not a named opening.
func (n *trieNode) lookup(uciMoves []string) (*Opening, bool) {
	current := n

	for _, move := range uciMoves {
		child, exists := current.children[move]
		if !exists {
			return nil, false
		}

		current = child
	}

	if current.opening == nil {
		return nil, false
	}

	return current.opening, true
}
