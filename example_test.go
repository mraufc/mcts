package mcts_test

import (
	"fmt"
	"math/rand"
	"runtime"
	"time"

	"github.com/mraufc/mcts"
	"github.com/mraufc/tictactoe/game"

	_ "github.com/mraufc/tictactoe/player"
)

// This example demonstrates 2 "AI" players backed with pure Monte Carlo Tree Search
// playing TicTacToe against each other.
// First scenario is played on a 3x3 board where 3 in a row vertically, horizontally or diagonally wins the game.
// With ideal moves, this game should always end in a draw.
// Second scenario is played on a 4x4 board where 3 in a row vertically, horizontally or diagonally wins the game.
// With ideal moves, first player (X) should always be able to generate a double attack and win in 3 first player moves
// or 5 total moves.
// This example is implemented to allow concurrent gameplays for convenience.
func Example() {
	numCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(numCPU)

	numGames := 500
	result := make(chan int, numGames)
	work := make(chan int, numGames)
	exit := make(chan int)
	for i := 0; i < numCPU; i++ {
		go NewWorker(3, 3, 3, 80*time.Millisecond, exit, work, result).Run()
	}

	for i := 0; i < numGames; i++ {
		work <- i
	}
	xWins, oWins, draws := 0, 0, 0
	for i := 0; i < numGames; i++ {
		winner := <-result
		if winner == 1 {
			xWins++
		} else if winner == 2 {
			oWins++
		} else {
			draws++
		}
	}

	for i := 0; i < numCPU; i++ {
		exit <- 0
	}

	fmt.Printf("3x3 Board with target 3 in a row number of games is equal to number of draws: %v\n", draws == numGames)

	// this is going to be a 4x4 board with target 3 in a row. X should always be able to create a double attack to win the game.
	result2 := make(chan int, numGames)
	work2 := make(chan int, numGames)
	exit2 := make(chan int)
	for i := 0; i < numCPU; i++ {
		go NewWorker(4, 4, 3, 180*time.Millisecond, exit2, work2, result2).Run()
	}

	for i := 0; i < numGames; i++ {
		work2 <- i
	}
	xWins, oWins, draws = 0, 0, 0
	for i := 0; i < numGames; i++ {
		winner := <-result2
		if winner == 1 {
			xWins++
		} else if winner == 2 {
			oWins++
		} else {
			draws++
		}
	}

	for i := 0; i < numCPU; i++ {
		exit2 <- 0
	}

	fmt.Printf("4x4 Board with target 3 in a row number of games is equal to number of X Player wins: %v\n", xWins == numGames)

	// Output:
	// 3x3 Board with target 3 in a row number of games is equal to number of draws: true
	// 4x4 Board with target 3 in a row number of games is equal to number of X Player wins: true
}

type Worker struct {
	p1     *Player
	p2     *Player
	en     *game.Engine
	exit   <-chan int
	work   <-chan int
	result chan<- int
}

func NewWorker(rows, columns, target int, searchDur time.Duration, exit, work <-chan int, result chan<- int) *Worker {
	engine, err := game.NewEngine(rows, columns, target)
	if err != nil {
		panic(err)
	}
	me1 := NewMoveEval(engine, rand.New(rand.NewSource(time.Now().UnixNano())))
	mg1 := &MoveGen{}

	search1 := mcts.New(me1, mg1)

	p1 := NewPlayer(search1, searchDur)
	me2 := NewMoveEval(engine, rand.New(rand.NewSource(time.Now().UnixNano())))
	mg2 := &MoveGen{}

	search2 := mcts.New(me2, mg2)
	p2 := NewPlayer(search2, searchDur)
	return &Worker{
		p1:     p1,
		p2:     p2,
		en:     engine,
		exit:   exit,
		work:   work,
		result: result,
	}
}

func (w *Worker) Run() {
	for {
		select {
		case <-w.exit:
			return
		case <-w.work:
			t, err := game.New(w.en, w.p1, w.p2)
			if err != nil {
				fmt.Println(err)
				return
			}
			for t.Play() {
			}
			_, winner := t.Result()
			w.result <- winner
		}
	}
}

// MoveGen implements mcts.Expander interface
type MoveGen struct {
}

func (mg *MoveGen) Expand(board [][]int, side int) []mcts.Move {
	res := make([]mcts.Move, 0)
	for i, row := range board {
		for j, v := range row {
			if v == 0 {
				m := &Move{i: i, j: j, side: side, eval: 0.0}
				res = append(res, m)
			}
		}
	}
	return res
}

// MoveEval implements mcts.Evaluator interface
type MoveEval struct {
	e *game.Engine
	r *rand.Rand
}

func NewMoveEval(e *game.Engine, r *rand.Rand) *MoveEval {
	return &MoveEval{e: e, r: r}
}

func (me *MoveEval) RandomMove(board [][]int, currentPlayerSide int) mcts.Move {
	empty := make([][]int, 0)
	for i, row := range board {
		for j, v := range row {
			if v == 0 {
				empty = append(empty, []int{i, j})
			}
		}
	}
	if len(empty) == 0 {
		return nil
	}
	ix := me.r.Intn(len(empty))
	return &Move{
		i:    empty[ix][0],
		j:    empty[ix][1],
		side: currentPlayerSide,
		eval: 0.0,
	}

}
func (me *MoveEval) ApplyMove(board [][]int, currentTurn int, m mcts.Move) (gameOver bool, winner int, err error) {
	mov := m.(*Move)
	gameOver, winner, err = me.e.Evaluate(board, currentTurn, mov.i, mov.j)
	if gameOver && winner != 0 && winner != currentTurn {
		return
	}
	board[mov.i][mov.j] = currentTurn
	return
}
func (me *MoveEval) NextPlayer(currentPlayerSide int) int {
	return 3 - currentPlayerSide
}

func (me *MoveEval) PrevPlayer(currentPlayerSide int) int {
	return 3 - currentPlayerSide
}

// Move implements mcts.Move interface
type Move struct {
	i, j int
	side int
	eval float64
}

func (m *Move) Eval() float64 {
	return m.eval
}

func (m *Move) PlayerSide() int {
	return m.side
}

// Player implements tictactoe/game.Player interface
type Player struct {
	name      string
	m         *mcts.MCTS
	searchDur time.Duration
}

func NewPlayer(m *mcts.MCTS, searchDur time.Duration) *Player {
	return &Player{m: m, searchDur: searchDur}
}

func (p *Player) Play(board [][]int, side int) (int, int) {
	move, _ := p.m.Search(board, side, p.searchDur, 0, 0)
	return move.(*Move).i, move.(*Move).j
}
func (p *Player) Done(winner int) {

}
func (p *Player) Name() string {
	return p.name
}
