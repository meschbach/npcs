package t3

import "fmt"

type PlayerError struct {
	WhichPlayer int
	Performing  string
	Underlying  error
}

func (p *PlayerError) Error() string {
	return fmt.Sprintf("encountered problem while player %d was %s: %s", p.WhichPlayer, p.Performing, p.Underlying)
}

func (p *PlayerError) Unwrap() error {
	return p.Underlying
}

type OffBoardError struct {
	To         Move
	MaxRows    int
	MaxColumns int
	Reason     string
}

func (m *OffBoardError) Error() string {
	return fmt.Sprintf("Move to (row: %d, column: %d); max(row: %d, columns: %d) invalid because %s", m.To.Row, m.To.Column, m.MaxRows, m.MaxColumns, m.Reason)
}
