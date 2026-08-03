package main

import (
	"chessbot-go/board"
	"chessbot-go/controller"
	"chessbot-go/game"
	"flag"
	"fmt"
	"os"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s <mode>\n\nModes:\n  self   AlphaBeta vs AlphaBeta (default)\n  fics   Play on FICS as guest\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	mode := "self"
	if flag.NArg() > 0 {
		mode = flag.Arg(0)
	}

	switch mode {
	case "self":
		white := func(b board.BitBoard) (board.Move, error) { return controller.AlphaBeta(b, board.White, 5) }
		black := func(b board.BitBoard) (board.Move, error) { return controller.AlphaBeta(b, board.Black, 5) }
		g := game.NewGame(white, black)
		r := g.Run()
		fmt.Printf("Result: %v\n", r)
	case "fics":
		r := controller.FicsGame(5)
		fmt.Printf("Result: %v\n", r)
	default:
		fmt.Fprintf(os.Stderr, "Unknown mode: %s\n", mode)
		flag.Usage()
		os.Exit(1)
	}
}
