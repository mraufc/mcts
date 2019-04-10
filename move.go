package mcts

// Move is a move that can be applied to a board.
// The Expander that lists the Moves to investigate can choose to return
// an evaluation for each Move, which will be backpropagated to parent
// nodes of the same player.
// Evaluation is recommended to be between -1.0 and 1.0 where -1.0 is a clearly losing
// evalution, 0.0 is a drawn evaluation and 1.0 is a clearly winning evaluation.
type Move interface {
	Eval() float64
}
