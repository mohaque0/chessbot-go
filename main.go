package main

import (
	"chessbot-go/board"
	"chessbot-go/controller"
	"chessbot-go/game"
	"flag"
	"fmt"
	"io"
	"os"
)

func main() {
	debug := flag.String("debug", "", "dump raw FICS data to file (use \"-\" for stderr)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [flags] <mode>\n\nModes:\n  self   AlphaBeta vs AlphaBeta (default)\n  fics   Play on FICS as guest\n\nFlags:\n", os.Args[0])
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
		var debugWr io.Writer
		if *debug == "-" {
			debugWr = os.Stderr
		} else if *debug != "" {
			f, err := os.Create(*debug)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to open debug file: %v\n", err)
				os.Exit(1)
			}
			defer f.Close()
			debugWr = f
		}
		r := controller.FicsGame(5, debugWr)
		fmt.Printf("Result: %v\n", r)
	default:
		fmt.Fprintf(os.Stderr, "Unknown mode: %s\n", mode)
		flag.Usage()
		os.Exit(1)
	}
}
