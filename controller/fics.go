package controller

import (
	"chessbot-go/board"
	"chessbot-go/fics"
	"chessbot-go/game"
	"fmt"
	"io"
	"math/rand"
	"strings"
	"time"
)

func FicsGame(depth uint, debugWr io.Writer) game.GameResult {
	client, err := fics.NewFicsClient(debugWr)
	if err != nil {
		fmt.Printf("Failed to connect to FICS: %v\n", err)
		return game.Draw
	}
	defer client.Close()

	fmt.Println("Connected to FICS. Waiting for login prompt...")

	if !waitForLoginPrompt(client) {
		fmt.Println("Did not receive login prompt.")
		return game.Draw
	}

	// Send "guest" at the login: prompt.
	client.Send <- fics.FicsSendText("guest")
	fmt.Println("Sent guest login.")

	// Wait for "Press return to enter the server as ..." prompt.
	username, ok := waitForGuestConfirm(client)
	if !ok {
		fmt.Println("Did not receive guest confirmation prompt.")
		return game.Draw
	}

	// Press return to enter.
	client.Send <- fics.FicsSendText("")
	fmt.Printf("Logged in as %s\n", username)

	// Configure output format.
	client.Send <- fics.FicsSendSetStyle{}
	client.Send <- fics.FicsSendSetNoWrap{}

	// Poll for sought games until we find one.
	var candidates []fics.FicsReceivedSoughtGame
	for {
		client.Send <- fics.FicsSendSought{}
		fmt.Println("Requesting sought games...")

		candidates = collectSoughtGames(client)
		if len(candidates) > 0 {
			break
		}
		fmt.Println("No standard or blitz games available, retrying in 5s...")
		time.Sleep(5 * time.Second)
	}

	chosen := candidates[rand.Intn(len(candidates))]
	fmt.Printf("Accepting game %d from %s\n", chosen.AdIdx, chosen.Username)
	client.Send <- fics.FicsSendPlay{GameID: chosen.AdIdx}

	// Wait for the first board to determine our color.
	first, ok := waitForBoard(client)
	if !ok {
		fmt.Println("Did not receive initial board.")
		return game.Draw
	}
	if first.gameEnd != nil {
		fmt.Printf("Game ended before it started: %s\n", first.gameEnd.Reason)
		return gameEndToResult(first.gameEnd)
	}
	firstBoard := first.board

	var ourColor board.Player
	if firstBoard.MyRelation == 1 {
		ourColor = board.White
	} else {
		ourColor = board.Black
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
			resp, ok := waitForBoard(client)
			if !ok {
				fmt.Println("Connection lost after sending move.")
				return game.Draw
			}
			if resp.gameEnd != nil {
				fmt.Printf("Game over: %s\n", resp.gameEnd.Reason)
				return gameEndToResult(resp.gameEnd)
			}
			if resp.illegalMove != nil {
				fmt.Printf("FICS rejected our move %s: %s\n", mv, resp.illegalMove.Text)
				fmt.Printf("Our board state:\n%s\n", currentBoard.String())
				return game.Draw
			}
			currentBoard = resp.board.Board
			mover = resp.board.Mover
			moveIdx++
			fmt.Printf("%d: %s played %s\n%s\n", moveIdx, ourColor, mv, currentBoard.String())
		} else {
			// Opponent's turn — wait for their move.
			resp, ok := waitForBoard(client)
			if !ok {
				fmt.Println("Connection lost waiting for opponent.")
				return game.Draw
			}
			if resp.gameEnd != nil {
				fmt.Printf("Game over: %s\n", resp.gameEnd.Reason)
				return gameEndToResult(resp.gameEnd)
			}
			if resp.illegalMove != nil {
				fmt.Printf("Unexpected illegal move message: %s\n", resp.illegalMove.Text)
				continue
			}
			currentBoard = resp.board.Board
			mover = resp.board.Mover
			moveIdx++

			moveStr := "unknown"
			if resp.board.LastMove != nil {
				moveStr = resp.board.LastMove.String()
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

func waitForLoginPrompt(client *fics.FicsClient) bool {
	for msg := range client.Recv {
		switch msg.(type) {
		case fics.FicsReceivedLoginPrompt:
			return true
		}
	}
	return false
}

func waitForGuestConfirm(client *fics.FicsClient) (string, bool) {
	username := ""
	for msg := range client.Recv {
		switch m := msg.(type) {
		case fics.FicsReceivedLoggedInAs:
			username = m.Username
		case fics.FicsReceivedRequestLogin:
			return username, true
		}
	}
	return "", false
}

type boardOrEnd struct {
	board       *fics.FicsReceivedBoard
	gameEnd     *fics.FicsReceivedGameEnd
	illegalMove *fics.FicsReceivedIllegalMove
}

func waitForBoard(client *fics.FicsClient) (boardOrEnd, bool) {
	for msg := range client.Recv {
		switch b := msg.(type) {
		case fics.FicsReceivedBoard:
			return boardOrEnd{board: &b}, true
		case fics.FicsReceivedGameEnd:
			return boardOrEnd{gameEnd: &b}, true
		case fics.FicsReceivedIllegalMove:
			return boardOrEnd{illegalMove: &b}, true
		}
	}
	return boardOrEnd{}, false
}

func collectSoughtGames(client *fics.FicsClient) []fics.FicsReceivedSoughtGame {
	var candidates []fics.FicsReceivedSoughtGame
	for msg := range client.Recv {
		switch sg := msg.(type) {
		case fics.FicsReceivedSoughtGame:
			if sg.GameType == fics.Standard || sg.GameType == fics.Blitz {
				candidates = append(candidates, sg)
			}
		case fics.FicsReceivedRequestUnknown:
			// "N ads displayed." marks the end of the sought list.
			// Skip all other unknown messages (MOTD, admin tells, etc.).
			if strings.Contains(sg.Text, "ad displayed") || strings.Contains(sg.Text, "ads displayed") {
				return candidates
			}
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

func gameEndToResult(e *fics.FicsReceivedGameEnd) game.GameResult {
	switch e.Result {
	case fics.WhiteWins:
		return game.WhiteWins
	case fics.BlackWins:
		return game.BlackWins
	default:
		return game.Draw
	}
}
