MCTS
=======

MCTS is a Monte Carlo Tree Search implementation.

Usage
=======

See [here](https://github.com/mraufc/mcts/blob/master/example_test.go) for an example that demonstrates 2 "AI" players backed with pure Monte Carlo Tree Search playing TicTacToe against each other.

First scenario is played on a 3x3 board where 3 in a row vertically, horizontally or diagonally wins the game. With ideal moves, this game should always end in a draw.

Second scenario is played on a 4x4 board where 3 in a row vertically, horizontally or diagonally wins the game. With ideal moves, first player (X) should always be able to generate a double attack and win in 3 first player (X) moves or 5 total moves.

Documentation
=======

See [here](https://godoc.org/github.com/mraufc/mcts) for GoDoc.