// Package mcts is a Monte Carlo Tree Search implementation.
// See https://en.wikipedia.org/wiki/Monte_Carlo_tree_search for more information.
package mcts

import (
	"math"
	"time"
)

// MCTS is the Monte Carlo Tree Search structure
type MCTS struct {
	ev Evaluator
	ex Expander
}

// New returns a new MCTS structure.
func New(ev Evaluator, ex Expander) *MCTS {
	return &MCTS{
		ev: ev,
		ex: ex,
	}
}

// Search searches the best Move for a side given a board for a limited duration.
// If maxIters is less than or equal to 0, the iteration count will only be limited by duration.
func (s *MCTS) Search(board [][]int, side int, duration time.Duration, maxDepth, maxIters int) (Move, int64) {
	t0 := time.Now()
	root := &treeNode{
		children: make([]*treeNode, 0),
		board:    board,
		depth:    0,
		side:     s.ev.PrevPlayer(side),
	}
	var node *treeNode
	iter := 0
	// run this loop at least once
	for iter == 0 || time.Since(t0) < duration {
		if maxIters > 0 && iter >= maxIters {
			break
		}
		iter++
		node = promisingNode(root)
		node.expand(s.ev, s.ex, maxDepth)
		node = firstChildOrItself(node)
		s.randomPlayOut(node)
		backpropagate(node)
	}

	return bestChild(root).move, root.visits
}

func (s *MCTS) randomPlayOut(n *treeNode) {
	if n.gameOver {
		return
	}
	currentTurn := s.ev.NextPlayer(n.side)

	board := copyBoard(n.board)
	for !n.gameOver {
		m := s.ev.RandomMove(board, currentTurn)
		if m == nil {
			break
		}
		gameOver, winner, err := s.ev.ApplyMove(board, currentTurn, m)
		if err != nil {
			panic(err)
		}
		if gameOver {
			n.winner = winner
			break
		}
		currentTurn = s.ev.NextPlayer(currentTurn)
	}
}

func (n *treeNode) expand(ev Evaluator, ex Expander, maxDepth int) {
	if n.gameOver {
		return
	}
	if maxDepth > 0 && n.depth >= maxDepth {
		return
	}
	nextPlayer := ev.NextPlayer(n.side)
	moves := ex.Expand(n.board, nextPlayer)
	for _, m := range moves {
		board := copyBoard(n.board)
		child := &treeNode{
			children: make([]*treeNode, 0),
			board:    board,
			depth:    n.depth + 1,
			move:     m,
			parent:   n,
			side:     nextPlayer,
		}
		n.children = append(n.children, child)
		gameOver, winner, err := ev.ApplyMove(board, nextPlayer, m)
		if err != nil {
			panic(err)
		}
		if gameOver {
			child.gameOver = true
			child.winner = winner
		}

		side := child.side
		for child != nil {
			child.visits++
			if child.side == side {
				child.winScore += m.Eval()
			} else {
				child.winScore -= m.Eval()
			}
			child = child.parent
		}
	}
}

func firstChildOrItself(n *treeNode) *treeNode {
	if len(n.children) == 0 || n.gameOver {
		return n
	}
	return n.children[0]
}

func bestChild(n *treeNode) *treeNode {
	if len(n.children) == 0 {
		panic("could not find any children")
	}
	res := n.children[0]
	maxVisits := res.visits
	for i := 1; i < len(n.children); i++ {
		ch := n.children[i]
		if ch.visits > maxVisits {
			res = ch
			maxVisits = ch.visits
		}
	}
	return res
}

func copyBoard(board [][]int) [][]int {
	res := make([][]int, len(board))
	for i, row := range board {
		res[i] = make([]int, len(row))
		for j, v := range row {
			res[i][j] = v
		}
	}
	return res
}

// treeNode is the search tree node.
// side is 1 for player 1 and 2 for player 2. For board games with more players,
// side can be 3 or more.
// winner is 0 for a draw, 1 for player 1 and 2 for player 2 and so on.
type treeNode struct {
	parent   *treeNode
	children []*treeNode
	side     int
	move     Move
	winner   int
	winScore float64
	visits   int64
	gameOver bool
	level    int
	board    [][]int
	depth    int
}

func promisingNode(n *treeNode) *treeNode {
	if n.gameOver {
		return n
	}
	res := n
	for len(res.children) > 0 {
		res = highestUCBChild(res)
	}
	return res
}

func highestUCBChild(n *treeNode) *treeNode {
	parentVisits := float64(n.visits)
	res := n.children[0]
	if res.visits == 0 {
		return res
	}
	visits := float64(res.visits)
	maxVal := (res.winScore / visits) + math.Sqrt2*math.Sqrt(math.Log(parentVisits)/visits)
	for i := 1; i < len(n.children); i++ {
		node := n.children[i]
		if node.visits == 0 {
			return node
		}
		visits = float64(node.visits)
		val := (node.winScore / visits) + math.Sqrt2*math.Sqrt(math.Log(parentVisits)/visits)
		if val > maxVal {
			maxVal = val
			res = node
		}
	}
	return res
}

func backpropagate(n *treeNode) {
	winner := n.winner
	for n != nil {
		n.visits++
		if winner != 0 {
			if winner == n.side {
				n.winScore += 1.0
			} else {
				n.winScore -= 1.0
			}
		}
		n = n.parent
	}
}
