package screenbuf

import (
	"bytes"
	"fmt"
	"io"
)

const esc = "\033["

var (
	clearLine = []byte(esc + "2K\r")
	moveUp    = []byte(esc + "1A")
	moveDown  = []byte(esc + "1B")
)

// ScreenBuf is a convenient way to write to terminal screens. It creates,
// clears and, moves up or down lines as needed to write the output to the
// terminal using ANSI escape codes.
type ScreenBuf struct {
	w      io.Writer
	buf    *bytes.Buffer
	reset  bool
	flush  bool
	cursor int
	height int
}

// New creates and initializes a new ScreenBuf struct.
func New(w io.Writer) *ScreenBuf {
	return &ScreenBuf{buf: &bytes.Buffer{}, w: w}
}

// Reset truncates the underlining buffer and marks all its previous lines to be
// cleared during the next Write.
func (s *ScreenBuf) Reset() {
	s.buf.Reset()
	s.reset = true
}

// Write writes a single line to the underlining buffer. If the ScreenBuf was
// previously reset, all previous lines are cleared and the output starts from
// the top. Lines with \r or \n will cause an error since they can interfere with the
// terminal ability to move between lines.
func (s *ScreenBuf) Write(b []byte) (int, error) {
	if bytes.ContainsAny(b, "\r\n") {
		return 0, fmt.Errorf("%q should not contain either \\r or \\n", b)
	}

	if s.reset {
		for i := 0; i < s.height; i++ {
			_, err := s.buf.Write(moveUp)
			if err != nil {
				return 0, err
			}
			_, err = s.buf.Write(clearLine)
			if err != nil {
				return 0, err
			}
		}
		s.cursor = 0
		s.height = 0
		s.reset = false
	}

	switch {
	case s.cursor == s.height:
		n, err := s.buf.Write(clearLine)
		if err != nil {
			return n, err
		}
		line := append(b, []byte("\n")...)
		n, err = s.buf.Write(line)
		if err != nil {
			return n, err
		}
		s.height++
		s.cursor++
		return n, nil
	case s.cursor < s.height:
		n, err := s.buf.Write(clearLine)
		if err != nil {
			return n, err
		}
		n, err = s.buf.Write(b)
		if err != nil {
			return n, err
		}
		n, err = s.buf.Write(moveDown)
		if err != nil {
			return n, err
		}
		s.cursor++
		return n, nil
	default:
		return 0, fmt.Errorf("Invalid write cursor position (%d) exceeded line height: %d", s.cursor, s.height)
	}
}

// Flush writes any buffered data to the underlying io.Writer, essentially clearing the console.
func (s *ScreenBuf) Flush() error {
	for i := s.cursor; i < s.height; i++ {
		if i < s.height {
			_, err := s.buf.Write(clearLine)
			if err != nil {
				return err
			}
		}
		_, err := s.buf.Write(moveDown)
		if err != nil {
			return err
		}
	}

	_, err := s.buf.WriteTo(s.w)
	if err != nil {
		return err
	}

	s.buf.Reset()

	for i := 0; i < s.height; i++ {
		_, err := s.buf.Write(moveUp)
		if err != nil {
			return err
		}
	}

	s.cursor = 0

	return nil
}

// WriteString is a convenient function to write a new line passing a string.
// Check ScreenBuf.Write() for a detailed explanation of the function behaviour.
func (s *ScreenBuf) WriteString(str string) (int, error) {
	return s.Write([]byte(str))
}
