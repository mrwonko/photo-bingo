package main

import "math/rand/v2"

type Goal struct {
	Name        string
	Description string
}

var freeGoal = Goal{Name: "Free Space", Description: "Automatically completed."}

const freeGoalIdx = -1

// DisplayBingoSpace is a denormalized [BingoSpace] for template rendering.
type DisplayBingoSpace struct {
	Goal      Goal
	Completed bool
	Image     string
	Locked    bool
}
type DisplayBingoRow [5]DisplayBingoSpace
type DisplayBingoBoard [5]DisplayBingoRow

func (board *DisplayBingoBoard) get(x, y int) *DisplayBingoSpace {
	return &(*board)[y][x]
}

func (board *BingoBoard) display() DisplayBingoBoard {
	var res DisplayBingoBoard
	for x := range 5 {
		for y := range 5 {
			src := board.get(x, y)
			isFreeGoal := src.GoalIdx == freeGoalIdx
			goal := freeGoal
			if !isFreeGoal {
				goal = options[src.GoalIdx]
			}
			*res.get(x, y) = DisplayBingoSpace{
				Goal:      goal,
				Completed: src.Completed,
				Locked:    isFreeGoal,
				Image:     src.Image,
			}
		}
	}
	return res
}

type BingoSpace struct {
	GoalIdx   int    `json:"ix"` // index into [options] or [freeGoalIdx]
	Completed bool   `json:"ok"`
	Image     string `json:"img"` // empty = none
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
	var goals [len(options)]int
	for i := range goals {
		goals[i] = i
	}
	rand.Shuffle(len(goals), func(i, j int) {
		goals[i], goals[j] = goals[j], goals[i]
	})
	var res BingoBoard
	i := 0
	for x := range 5 {
		for y := range 5 {
			space := res.get(x, y)
			if x == 2 && y == 2 {
				space.GoalIdx = freeGoalIdx
				space.Completed = true
			} else {
				space.GoalIdx = goals[i]
				i++
			}
		}
	}
	return res
}
