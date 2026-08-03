package fics

import (
	"strings"
	"testing"
)

func TestParseSoughtLine(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantOK   bool
		wantAd   uint
		wantUser string
		wantType GameType
	}{
		{
			name:     "rated blitz with rating",
			input:    `  6 1803 Bersal              5   2 unrated blitz      [black]     0-9999 `,
			wantOK:   true,
			wantAd:   6,
			wantUser: "Bersal",
			wantType: Blitz,
		},
		{
			name:     "unrated blitz guest no color",
			input:    ` 42 ++++ GuestYLFR           3   0 unrated blitz                  0-9999 `,
			wantOK:   true,
			wantAd:   42,
			wantUser: "GuestYLFR",
			wantType: Blitz,
		},
		{
			name:     "computer player with (C)",
			input:    ` 64 1968 GriffyJr(C)         2  12 unrated blitz                  0-9999 `,
			wantOK:   true,
			wantAd:   64,
			wantUser: "GriffyJr",
			wantType: Blitz,
		},
		{
			name:     "standard game",
			input:    ` 20 ++++ Superepi            15   0 unrated standard m             0-9999 `,
			wantOK:   true,
			wantAd:   20,
			wantUser: "Superepi",
			wantType: Standard,
		},
		{
			name:     "lightning game",
			input:    ` 26 ++++ GuestJPGM           1   0 unrated lightning               0-9999 `,
			wantOK:   true,
			wantAd:   26,
			wantUser: "GuestJPGM",
			wantType: Lightning,
		},
		{
			name:     "ads displayed line",
			input:    `9 ads displayed.`,
			wantOK:   false,
		},
		{
			name:     "1 ad displayed line",
			input:    `1 ad displayed.`,
			wantOK:   false,
		},
		{
			name:     "empty string",
			input:    ``,
			wantOK:   false,
		},
		{
			name:     "random MOTD text",
			input:    `Welcome to FICS - the Free Internet Chess Server.`,
			wantOK:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := parseSoughtLine(tt.input)
			if ok != tt.wantOK {
				t.Fatalf("parseSoughtLine(%q) ok = %v, want %v", tt.input, ok, tt.wantOK)
			}
			if !ok {
				return
			}
			if got.AdIdx != tt.wantAd {
				t.Errorf("AdIdx = %d, want %d", got.AdIdx, tt.wantAd)
			}
			if got.Username != tt.wantUser {
				t.Errorf("Username = %q, want %q", got.Username, tt.wantUser)
			}
			if got.GameType != tt.wantType {
				t.Errorf("GameType = %v, want %v", got.GameType, tt.wantType)
			}
		})
	}
}

func TestReadFromFicsSoughtGames(t *testing.T) {
	// Simulate exact lines the telnet client delivers from a real FICS
	// "sought" response. The telnet client now strips \r and \n before
	// delivering lines.
	rawLines := []string{
		"fics%   6 1803 Bersal              5   2 unrated blitz      [black]     0-9999 ",
		" 42 ++++ GuestYLFR           3   0 unrated blitz                  0-9999 ",
		" 50 1457 Shlakh              3  12 unrated blitz                  0-9999 ",
		" 64 1968 GriffyJr(C)         2  12 unrated blitz                  0-9999 ",
		"9 ads displayed.",
	}

	src := make(chan string, len(rawLines))
	for _, line := range rawLines {
		src <- line
	}
	close(src)

	dst := make(chan FicsMessageReceived, 20)
	readFromFics(src, dst)

	var sought []FicsReceivedSoughtGame
	var unknown []FicsReceivedRequestUnknown
	for msg := range dst {
		switch m := msg.(type) {
		case FicsReceivedSoughtGame:
			sought = append(sought, m)
		case FicsReceivedRequestUnknown:
			unknown = append(unknown, m)
		}
	}

	if len(sought) != 4 {
		t.Fatalf("got %d sought games, want 4; unknown messages: %v", len(sought), formatUnknowns(unknown))
	}

	// Verify first game (was prefixed with "fics%").
	if sought[0].AdIdx != 6 || sought[0].Username != "Bersal" {
		t.Errorf("first game: got ad=%d user=%q, want ad=6 user=Bersal", sought[0].AdIdx, sought[0].Username)
	}
	// Verify second game (no fics% prefix, leading spaces).
	if sought[1].AdIdx != 42 || sought[1].Username != "GuestYLFR" {
		t.Errorf("second game: got ad=%d user=%q, want ad=42 user=GuestYLFR", sought[1].AdIdx, sought[1].Username)
	}
}

func TestReadFromFicsGameEnd(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantResult GameEndResult
		wantReason string
	}{
		{
			name:       "resignation white wins",
			input:      `{Game 26 (GuestFRZF vs. GuestNPXS) GuestNPXS resigns} 1-0`,
			wantResult: WhiteWins,
			wantReason: "GuestNPXS resigns",
		},
		{
			name:       "resignation black wins",
			input:      `{Game 42 (Alice vs. Bob) Alice resigns} 0-1`,
			wantResult: BlackWins,
			wantReason: "Alice resigns",
		},
		{
			name:       "checkmate",
			input:      `{Game 10 (Foo vs. Bar) Bar checkmated} 1-0`,
			wantResult: WhiteWins,
			wantReason: "Bar checkmated",
		},
		{
			name:       "draw",
			input:      `{Game 5 (X vs. Y) Game drawn by agreement} 1/2-1/2`,
			wantResult: DrawResult,
			wantReason: "Game drawn by agreement",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := make(chan string, 1)
			src <- tt.input
			close(src)

			dst := make(chan FicsMessageReceived, 5)
			readFromFics(src, dst)

			msg := <-dst
			ge, ok := msg.(FicsReceivedGameEnd)
			if !ok {
				t.Fatalf("expected FicsReceivedGameEnd, got %T", msg)
			}
			if ge.Result != tt.wantResult {
				t.Errorf("Result = %v, want %v", ge.Result, tt.wantResult)
			}
			if ge.Reason != tt.wantReason {
				t.Errorf("Reason = %q, want %q", ge.Reason, tt.wantReason)
			}
		})
	}
}

func formatUnknowns(msgs []FicsReceivedRequestUnknown) string {
	var parts []string
	for _, m := range msgs {
		parts = append(parts, m.Text)
	}
	return strings.Join(parts, " | ")
}
