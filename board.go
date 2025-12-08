package main

type BingoSpace struct {
	Name      string
	Completed bool
	Image     string // empty = none
}

type BingoRow [5]BingoSpace
type BingoBoard [5]BingoRow

func (board *BingoBoard) get(x, y int) *BingoSpace {
	return &(*board)[y][x]
}

func (board *BingoBoard) score() int {
	var (
		rows  [5]int
		cols  [5]int
		diags [2]int
	)
	for x := range 5 {
		for y := range 5 {
			if board.get(x, y).Completed {
				cols[x]++
				rows[y]++
			}
		}
		if board.get(x, x).Completed {
			diags[0]++
		}
		if board.get(x, 4-x).Completed {
			diags[1]++
		}
	}
	res := 0
	for i := range 5 {
		if rows[i] == 5 {
			res++
		}
		if cols[i] == 5 {
			res++
		}
	}
	for i := range 2 {
		if diags[i] == 5 {
			res++
		}
	}
	return res
}
