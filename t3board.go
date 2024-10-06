package npcs

import "errors"

type T3Move struct {
	Player int
	Row    int
	Column int
}

type T3Board struct {
	cells      [][]int
	maxRows    int
	maxColumns int
}

func NewT3Board() *T3Board {
	rows := 3
	columns := 3
	board := make([][]int, rows)
	for i := range board {
		board[i] = make([]int, columns)
	}
	return &T3Board{
		cells:      board,
		maxRows:    3,
		maxColumns: 3,
	}
}

func (t *T3Board) place(where T3Move) (occupied bool, err error) {
	if where.Row > t.maxRows {
		return false, errors.New("out of range")
	}
	if where.Column > t.maxColumns {
		return false, errors.New("out of range")
	}
	if where.Player == 0 {
		return false, errors.New("side 0 is reserved")
	}

	current := t.cells[where.Row][where.Column]
	if current != 0 {
		return true, nil
	}
	t.cells[where.Row][where.Column] = where.Player
	return false, nil
}

func (t *T3Board) completed(player int) bool {
	if t.rowWin(player) {
		return true
	}
	if t.columnWin(player) {
		return true
	}
	// Does the top left to bottom right contain a win?
	if t.backSlashWin(player) {
		return true
	}

	return t.forwardSlashWin(player)
}

func (t *T3Board) rowWin(player int) bool {
	for _, row := range t.cells {
		runCount := 0
		for _, cell := range row {
			if cell != player {
				break
			}
			runCount++
		}
		if runCount == 3 {
			return true
		}
	}
	return false
}

func (t *T3Board) forwardSlashWin(player int) bool {
	for index := 0; index < t.maxColumns; index++ {
		row := t.maxRows - index - 1
		column := index
		if t.cells[row][column] != player {
			return false
		}
	}
	return true
}

func (t *T3Board) backSlashWin(player int) bool {
	runCount := 0
	for index := 0; index < t.maxRows; index++ {
		if t.cells[index][index] != player {
			break
		}
		runCount++
	}
	return runCount == t.maxRows
}

func (t *T3Board) columnWin(player int) bool {
	for columnIndex := 0; columnIndex < t.maxColumns; columnIndex++ {
		runCount := 0
		for _, row := range t.cells {
			if row[columnIndex] != player {
				break
			}
			runCount++
		}
		if runCount == t.maxColumns {
			return true
		}
	}
	return false
}
