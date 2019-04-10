package mcts

// Expander should be able to return a list possible preferably legal moves to add to the tree as leaves given
// a board and current side.
type Expander interface {
	Expand(board [][]int, side int) []Move
}
