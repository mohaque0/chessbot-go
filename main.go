package main

import (
	"chessbot-go/game"
	"fmt"
)

func main() {
	g := game.NewGame()
	r := g.Run()
	fmt.Printf("Result: %v\n", r)
}
