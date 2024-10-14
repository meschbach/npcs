package tui

import (
	"errors"
	"fmt"
	"github.com/meschbach/npcs/t3"
	"strings"
)

func renderBoardAsString(b *t3.Board) (string, error) {
	builder := strings.Builder{}

	var errs []error
	places := make([][]string, 3)
	for row := 0; row < 3; row++ {
		places[row] = make([]string, 3)
		for col := 0; col < 3; col++ {
			who, err := b.Occupied(t3.Move{Row: row, Column: col})
			var sym string
			if err != nil {
				errs = append(errs, err)
				sym = "e"
			} else {
				sym = renderSideAsSymbol(who)
			}
			builder.WriteString(fmt.Sprintf(" %s ", sym))
			if col != 2 {
				builder.WriteString("|")
			} else {
				builder.WriteString("\n")
			}
		}
		if row != 2 {
			builder.WriteString(strings.Repeat("-", 11))
			builder.WriteString("\n")
		}
	}
	return builder.String(), errors.Join(errs...)
}

func renderSideAsSymbol(side int) string {
	switch side {
	case 0:
		return " "
	case 1:
		return "X"
	case 2:
		return "O"
	default:
		return fmt.Sprintf("%d", side)
	}
}
