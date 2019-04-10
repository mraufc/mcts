package mcts

// Evaluator should be able to perform the following actions:
//
// 1. Apply a move to a board and return the evaluation result.
// 2. Return the next player, given the current player side.
// 3. Return a random valid move, given a board and a player side.
type Evaluator interface {
	RandomMove(board [][]int, currentPlayerSide int) Move
	ApplyMove(board [][]int, currentPlayerSide int, m Move) (gameOver bool, winner int, err error)
	NextPlayer(currentPlayerSide int) int
	PrevPlayer(currentPlayerSide int) int
}
