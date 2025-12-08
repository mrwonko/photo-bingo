package main

import "github.com/mrwonko/photo-bingo/muxval"

type GameState struct {
	Players map[PlayerName]PlayerState
}

var gameState muxval.MuxVal[GameState]

type PlayerState struct {
	Password InsecurePlaintextPassword
	Approved bool
	Board    BingoBoard
}

type PlayerName string
type InsecurePlaintextPassword string
