package main

import "math/rand/v2"

type Goal struct {
	Name        string
	Description string
}

type BingoSpace struct {
	Goal      Goal
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

func generateBoard() BingoBoard {
	goals := options
	copy(goals[:], options[:])
	rand.Shuffle(len(goals), func(i, j int) {
		goals[i], goals[j] = goals[j], goals[i]
	})
	var res BingoBoard
	i := 0
	for x := range 5 {
		for y := range 5 {
			space := res.get(x, y)
			if x == 2 && y == 2 {
				space.Goal = Goal{Name: "Free Space", Description: "Automatically completed."}
				space.Completed = true
			} else {
				space.Goal = goals[i]
				i++
			}
		}
	}
	return res
}
