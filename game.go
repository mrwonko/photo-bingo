package main

type GameState struct {
	Players map[PlayerName]PlayerState
}

type PlayerState struct {
	Password InsecurePlaintextPassword
	Approved bool
	Board    BingoBoard
}

type PlayerName string
type InsecurePlaintextPassword string
