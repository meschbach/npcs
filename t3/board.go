package t3

import "errors"

type Move struct {
	Player int
	Row    int
	Column int
}

type Board struct {
	cells      [][]int
	maxRows    int
	maxColumns int
}

func NewBoard() *Board {
	rows := 3
	columns := 3
	board := make([][]int, rows)
	for i := range board {
		board[i] = make([]int, columns)
	}
	return &Board{
		cells:      board,
		maxRows:    3,
		maxColumns: 3,
	}
}

func (t *Board) Place(where Move) (occupied bool, err error) {
	if err := t.validOrError(where); err != nil {
		return false, err
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

func (t *Board) validOrError(where Move) error {
	if where.Row >= t.maxRows {
		return &OffBoardError{
			To:         where,
			MaxRows:    t.maxRows,
			MaxColumns: t.maxColumns,
			Reason:     "row",
		}
	}
	if where.Column >= t.maxColumns {
		return &OffBoardError{
			To:         where,
			MaxRows:    t.maxRows,
			MaxColumns: t.maxColumns,
			Reason:     "column",
		}
	}
	return nil
}

func (t *Board) Occupied(where Move) (player int, err error) {
	if err := t.validOrError(where); err != nil {
		return 0, err
	}
	return t.cells[where.Row][where.Column], nil
}

func (t *Board) completed(player int) bool {
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

func (t *Board) rowWin(player int) bool {
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

func (t *Board) forwardSlashWin(player int) bool {
	for index := 0; index < t.maxColumns; index++ {
		row := t.maxRows - index - 1
		column := index
		if t.cells[row][column] != player {
			return false
		}
	}
	return true
}

func (t *Board) backSlashWin(player int) bool {
	runCount := 0
	for index := 0; index < t.maxRows; index++ {
		if t.cells[index][index] != player {
			break
		}
		runCount++
	}
	return runCount == t.maxRows
}

func (t *Board) columnWin(player int) bool {
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
