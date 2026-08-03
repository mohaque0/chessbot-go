package controller

import (
	"chessbot-go/board"
	"chessbot-go/fics"
	"chessbot-go/game"
	"fmt"
	"math/rand"
)

func FicsGame(depth uint) game.GameResult {
	client, err := fics.NewFicsClient()
	if err != nil {
		fmt.Printf("Failed to connect to FICS: %v\n", err)
		return game.Draw
	}
	defer client.Close()

	fmt.Println("Connected to FICS. Waiting for login prompt...")

	if !waitForLogin(client) {
		fmt.Println("Did not receive login prompt.")
		return game.Draw
	}

	// Log in as guest.
	client.Send <- fics.FicsSendText("")
	fmt.Println("Logged in as guest.")

	// Configure output format.
	client.Send <- fics.FicsSendSetStyle{}
	client.Send <- fics.FicsSendSetNoWrap{}

	// Request sought games.
	client.Send <- fics.FicsSendSought{}
	fmt.Println("Requesting sought games...")

	candidates := collectSoughtGames(client)
	if len(candidates) == 0 {
		fmt.Println("No standard or blitz games available.")
		return game.Draw
	}

	chosen := candidates[rand.Intn(len(candidates))]
	fmt.Printf("Accepting game %d from %s\n", chosen.AdIdx, chosen.Username)
	client.Send <- fics.FicsSendPlay{GameID: chosen.AdIdx}

	// Wait for the first board to determine our color.
	firstBoard, ok := waitForBoard(client)
	if !ok {
		fmt.Println("Did not receive initial board.")
		return game.Draw
	}

	// The mover in the first board tells us whose turn it is.
	// If it's our turn, we're that color.
	var ourColor board.Player
	if firstBoard.LastMove == nil {
		// No move has been made yet — we are white.
		ourColor = board.White
	} else {
		// A move was made; the mover is us.
		ourColor = firstBoard.Mover
	}
	fmt.Printf("Playing as %s\n", ourColor)

	currentBoard := firstBoard.Board
	mover := firstBoard.Mover

	moveIdx := 0
	fmt.Printf("%d\n%s\n", moveIdx, currentBoard.String())

	for {
		if mover == ourColor {
			// Our turn — compute with AlphaBeta.
			mv, err := AlphaBeta(currentBoard, ourColor, depth)
			if err != nil {
				fmt.Printf("No moves available: %v\n", err)
				return game.Draw
			}

			client.Send <- fics.FicsSendMove{Move: mv}
			fmt.Printf("Sent move: %s\n", mv)

			// Wait for server confirmation via board update.
			boardMsg, ok := waitForBoard(client)
			if !ok {
				fmt.Println("Connection lost after sending move.")
				return game.Draw
			}
			currentBoard = boardMsg.Board
			mover = boardMsg.Mover
			moveIdx++
			fmt.Printf("%d: %s played %s\n%s\n", moveIdx, ourColor, mv, currentBoard.String())
		} else {
			// Opponent's turn — wait for their move.
			boardMsg, ok := waitForBoard(client)
			if !ok {
				fmt.Println("Connection lost waiting for opponent.")
				return game.Draw
			}
			currentBoard = boardMsg.Board
			mover = boardMsg.Mover
			moveIdx++

			moveStr := "unknown"
			if boardMsg.LastMove != nil {
				moveStr = boardMsg.LastMove.String()
			}
			fmt.Printf("%d: %s played %s\n%s\n", moveIdx, ourColor.Other(), moveStr, currentBoard.String())
		}

		// Check for game end.
		if len(currentBoard.GetMoves(mover)) == 0 {
			if currentBoard.IsInCheck(mover) {
				if mover == ourColor {
					fmt.Println("We are checkmated.")
				} else {
					fmt.Println("Opponent is checkmated!")
				}
				return playerToResult(mover.Other())
			}
			fmt.Println("Stalemate.")
			return game.Draw
		}
	}
}

func waitForLogin(client *fics.FicsClient) bool {
	for msg := range client.Recv {
		switch msg.(type) {
		case fics.FicsReceivedRequestLogin:
			return true
		}
	}
	return false
}

func waitForBoard(client *fics.FicsClient) (fics.FicsReceivedBoard, bool) {
	for msg := range client.Recv {
		switch b := msg.(type) {
		case fics.FicsReceivedBoard:
			return b, true
		}
	}
	return fics.FicsReceivedBoard{}, false
}

func collectSoughtGames(client *fics.FicsClient) []fics.FicsReceivedSoughtGame {
	var candidates []fics.FicsReceivedSoughtGame
	for msg := range client.Recv {
		switch sg := msg.(type) {
		case fics.FicsReceivedSoughtGame:
			if sg.GameType == fics.Standard || sg.GameType == fics.Blitz {
				candidates = append(candidates, sg)
			}
		default:
			// Non-sought message means the sought list is done.
			return candidates
		}
	}
	return candidates
}

func playerToResult(p board.Player) game.GameResult {
	if p == board.White {
		return game.WhiteWins
	}
	return game.BlackWins
}
