package fics

import (
	"bytes"
	"testing"
)

func TestStripTelnetIAC_NoIAC(t *testing.T) {
	input := []byte("hello world\r\n")
	got := stripTelnetIAC(input)
	if !bytes.Equal(got, input) {
		t.Errorf("expected %q, got %q", input, got)
	}
}

func TestStripTelnetIAC_EscapedFF(t *testing.T) {
	// IAC IAC (0xFF 0xFF) should produce a single 0xFF byte.
	input := []byte{0xFF, 0xFF}
	got := stripTelnetIAC(input)
	expected := []byte{0xFF}
	if !bytes.Equal(got, expected) {
		t.Errorf("expected %x, got %x", expected, got)
	}
}

func TestStripTelnetIAC_WillCommand(t *testing.T) {
	// IAC WILL option (3 bytes) should be stripped entirely.
	input := []byte{'a', 0xFF, 0xFB, 0x01, 'b'}
	got := stripTelnetIAC(input)
	expected := []byte{'a', 'b'}
	if !bytes.Equal(got, expected) {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestStripTelnetIAC_WontCommand(t *testing.T) {
	input := []byte{'x', 0xFF, 0xFC, 0x03, 'y'}
	got := stripTelnetIAC(input)
	expected := []byte{'x', 'y'}
	if !bytes.Equal(got, expected) {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestStripTelnetIAC_DoCommand(t *testing.T) {
	input := []byte{0xFF, 0xFD, 0x18, 'h', 'i'}
	got := stripTelnetIAC(input)
	expected := []byte{'h', 'i'}
	if !bytes.Equal(got, expected) {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestStripTelnetIAC_DontCommand(t *testing.T) {
	input := []byte{'a', 0xFF, 0xFE, 0x01}
	got := stripTelnetIAC(input)
	expected := []byte{'a'}
	if !bytes.Equal(got, expected) {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestStripTelnetIAC_TwoByteCommand(t *testing.T) {
	// IAC GA (Go Ahead, 0xF9) is a 2-byte command.
	input := []byte{'a', 0xFF, 0xF9, 'b'}
	got := stripTelnetIAC(input)
	expected := []byte{'a', 'b'}
	if !bytes.Equal(got, expected) {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestStripTelnetIAC_MultipleSequences(t *testing.T) {
	input := []byte{
		0xFF, 0xFD, 0x18, // IAC DO 0x18
		'h', 'e', 'l', 'l', 'o',
		0xFF, 0xFB, 0x01, // IAC WILL 0x01
		' ',
		0xFF, 0xF9, // IAC GA
		'!',
	}
	got := stripTelnetIAC(input)
	expected := []byte("hello !")
	if !bytes.Equal(got, expected) {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestStripTelnetIAC_EmbeddedInFICSLine(t *testing.T) {
	// Simulate a FICS sought line with IAC sequences mixed in.
	line := []byte("  6 ++++ captainlonely       7   7 unrated blitz                  0-9999\r\n")
	input := make([]byte, 0)
	input = append(input, 0xFF, 0xF9) // IAC GA at start
	input = append(input, line...)
	got := stripTelnetIAC(input)
	if !bytes.Equal(got, line) {
		t.Errorf("expected %q, got %q", line, got)
	}
}

func TestStripTelnetIAC_TrailingIACAtEnd(t *testing.T) {
	// IAC at the very end with no following byte should be kept as-is.
	input := []byte{'a', 0xFF}
	got := stripTelnetIAC(input)
	expected := []byte{'a', 0xFF}
	if !bytes.Equal(got, expected) {
		t.Errorf("expected %x, got %x", expected, got)
	}
}
