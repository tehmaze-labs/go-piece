package parser

import (
	"log"

	"github.com/tehmaze-labs/go-piece/buffer"
	"github.com/tehmaze-labs/go-piece/calc"
)

// Cursor Character Absolute
func (p *ANSI) parseCHA(s *ANSISequence) (err error) {
	x := 0
	if s.Len() > 0 {
		x = s.Int(0) - 1
	}
	p.buffer.Cursor.X = calc.MaxInt(0, x)
	return
}

// Cursor Next Line
func (p *ANSI) parseCNL(s *ANSISequence) (err error) {
	y := 1
	if s.Len() > 0 {
		y = s.Int(0)
	}
	p.buffer.Cursor.X = 0
	p.buffer.Cursor.Down(y)
	return
}

// Cursor Preceding Line
func (p *ANSI) parseCPL(s *ANSISequence) (err error) {
	y := 1
	if s.Len() > 0 {
		y = s.Int(0)
	}
	p.buffer.Cursor.X = 0
	p.buffer.Cursor.Up(y)
	return
}

// Cursor Left
func (p *ANSI) parseCUB(s *ANSISequence) (err error) {
	x := 1
	if s.Len() > 0 {
		x = s.Int(0)
	}
	p.buffer.Cursor.Left(x)
	return
}

// Cursor Down
func (p *ANSI) parseCUD(s *ANSISequence) (err error) {
	y := 1
	if s.Len() > 0 {
		y = s.Int(0)
	}
	p.buffer.Cursor.Down(y)
	return
}

// Cursor Right
func (p *ANSI) parseCUF(s *ANSISequence) (err error) {
	x := 1
	if s.Len() > 0 {
		x = s.Int(0)
	}
	p.buffer.Cursor.Right(x)
	return
}

// Cursor Position
func (p *ANSI) parseCUP(s *ANSISequence) (err error) {
	y := 0
	x := 0
	switch s.Len() {
	case 2:
		y = s.Int(0) - 1
		x = s.Int(1) - 1
	case 1:
		y = s.Int(0) - 1
	}
	p.buffer.Cursor.Goto(x, y)
	return
}

func (p *ANSI) parseCUU(s *ANSISequence) (err error) {
	return
}

// Erase Display
func (p *ANSI) parseED(s *ANSISequence) (err error) {
	i := s.Int(0)

	switch i {
	case 0: // From cursor to EOF
		o := p.buffer.Cursor.Offset(p.buffer.Width)
		p.buffer.ClearFrom(o)
	case 1: // From cursor to start
		o := p.buffer.Cursor.Offset(p.buffer.Width)
		p.buffer.ClearTo(o)
	default: // Entire buffer
		p.buffer.Clear()
	}

	return
}

// Erase Line
func (p *ANSI) parseEL(s *ANSISequence) (err error) {
	i := s.Int(0)
	var o, e int
	switch i {
	case 0: // To EOL
		o = (p.buffer.Width * (p.buffer.Cursor.Y)) + p.buffer.Cursor.X
		e = (p.buffer.Width * (p.buffer.Cursor.Y + 1)) - 1
	case 1: // To BOL
		o = (p.buffer.Width * (p.buffer.Cursor.Y - 1)) + 1
		e = (p.buffer.Width * (p.buffer.Cursor.Y)) + p.buffer.Cursor.X
	case 2: // From BOL to EOL
		o = (p.buffer.Width * (p.buffer.Cursor.Y - 1)) + 1
		e = (p.buffer.Width * (p.buffer.Cursor.Y + 1)) - 1
	}

	o = calc.MaxInt(o, 0)
	e = calc.MinInt(e, p.buffer.Len())

	for i = o; i < e; i++ {
		p.buffer.ClearAt(i)
	}

	return
}

// Insert Line
func (p *ANSI) parseIL(s *ANSISequence) (err error) {
	i := 1
	if s.Len() > 0 {
		i = s.Int(0)
	}
	o := p.buffer.Width * p.buffer.Normalize().Cursor.Y
	for ; i > 0; i-- {
		p.buffer.Insert(o, p.buffer.Width)
	}
	return
}

func (p *ANSI) parseSGR(s *ANSISequence) (err error) {
	for _, n := range s.Ints() {
		switch n {
		// ECMA-48 standard codes
		case 0: // Default rendition
			p.buffer.Cursor.ResetAttrib()
		case 1: // Bold
			p.buffer.Cursor.Attrib |= buffer.ATTRIB_BOLD
		case 2: // Faint
			p.buffer.Cursor.Attrib |= buffer.ATTRIB_FAINT
		case 3: // Italicized
			p.buffer.Cursor.Attrib |= buffer.ATTRIB_ITALICS
		case 4: // Underlined
			p.buffer.Cursor.Attrib &^= buffer.ATTRIB_UNDERLINE_DOUBLE
			p.buffer.Cursor.Attrib |= buffer.ATTRIB_UNDERLINE
		case 5, 6: // Blink
			p.buffer.Cursor.Attrib |= buffer.ATTRIB_BLINK
		case 7: // Negative
			p.buffer.Cursor.Attrib |= buffer.ATTRIB_NEGATIVE
		case 8: // Concealed characters
			p.buffer.Cursor.Attrib |= buffer.ATTRIB_CONCEAL
		case 9: // Crossed-out
			p.buffer.Cursor.Attrib |= buffer.ATTRIB_CROSS_OUT
		case 10, 11, 12, 13, 14, 15, 16, 17, 18, 19:
			p.buffer.Cursor.Font = n - 10
		case 20: // Fraktur (Gothic)
			p.buffer.Cursor.Attrib |= buffer.ATTRIB_GOTHIC
		case 21: // Doubly underlined
			p.buffer.Cursor.Attrib &^= buffer.ATTRIB_UNDERLINE
			p.buffer.Cursor.Attrib |= buffer.ATTRIB_UNDERLINE_DOUBLE
		case 22: // Neither bold nor faint
			p.buffer.Cursor.Attrib &^= buffer.ATTRIB_BOLD
			p.buffer.Cursor.Attrib &^= buffer.ATTRIB_FAINT
		case 23: // Neither italicized nor fraktur
			p.buffer.Cursor.Attrib &^= buffer.ATTRIB_ITALICS
			p.buffer.Cursor.Attrib &^= buffer.ATTRIB_GOTHIC
		case 24: // Not underlined
			p.buffer.Cursor.Attrib &^= buffer.ATTRIB_UNDERLINE
			p.buffer.Cursor.Attrib &^= buffer.ATTRIB_UNDERLINE_DOUBLE
		case 25: // Not blinking
			p.buffer.Cursor.Attrib &^= buffer.ATTRIB_BLINK
		case 26: // Reserved
		case 27: // Positive
			p.buffer.Cursor.Attrib &^= buffer.ATTRIB_NEGATIVE
		case 28: // Revealed
			p.buffer.Cursor.Attrib &^= buffer.ATTRIB_CONCEAL
		case 29: // Not crossed out
			p.buffer.Cursor.Attrib &^= buffer.ATTRIB_CROSS_OUT
		case 30, 31, 32, 33, 34, 35, 36, 37:
			p.buffer.Cursor.Color = n - 30
		case 38: // Reserved (TODO 24 bit ANSi)
		case 39: // Default display colour
			p.buffer.Cursor.Color = buffer.TILE_DEFAULT_COLOR
		case 40, 41, 42, 43, 44, 45, 46, 47:
			p.buffer.Cursor.Background = n - 40
		case 48: // Reserved (TODO 24 bit ANSi)
		case 49: // Default background colour
			p.buffer.Cursor.Background = buffer.TILE_DEFAULT_BACKGROUND
		case 50: // Reserved (cancels 26)
		case 51: // Framed
			p.buffer.Cursor.Attrib |= buffer.ATTRIB_FRAME
		case 52: // Encircled
			p.buffer.Cursor.Attrib |= buffer.ATTRIB_ENCIRCLE
		case 53: // Overlined
			p.buffer.Cursor.Attrib |= buffer.ATTRIB_OVERLINE
		case 54: // Not framed nor encircled
			p.buffer.Cursor.Attrib &^= buffer.ATTRIB_FRAME
			p.buffer.Cursor.Attrib &^= buffer.ATTRIB_ENCIRCLE
		case 55: // Not overlined
			p.buffer.Cursor.Attrib &^= buffer.ATTRIB_OVERLINE
		case 56, 57, 58, 59: // Reserved
		case 60: // Ideogram underline
			p.buffer.Cursor.Attrib |= buffer.ATTRIB_IDEOGRAM_UNDERLINE
		case 61: // Ideogram double underline
			p.buffer.Cursor.Attrib |= buffer.ATTRIB_IDEOGRAM_UNDERLINE_DOUBLE
		case 62: // Ideogram overline
			p.buffer.Cursor.Attrib |= buffer.ATTRIB_IDEOGRAM_OVERLINE
		case 63: // Ideogram double overline
			p.buffer.Cursor.Attrib |= buffer.ATTRIB_IDEOGRAM_OVERLINE_DOUBLE
		case 64: // Ideogram stress marking
			p.buffer.Cursor.Attrib |= buffer.ATTRIB_IDEOGRAM_STRESS_MARKING
		case 65: // Cancels 60..64
			p.buffer.Cursor.Attrib &^= buffer.ATTRIB_IDEOGRAM_UNDERLINE
			p.buffer.Cursor.Attrib &^= buffer.ATTRIB_IDEOGRAM_UNDERLINE_DOUBLE
			p.buffer.Cursor.Attrib &^= buffer.ATTRIB_IDEOGRAM_OVERLINE
			p.buffer.Cursor.Attrib &^= buffer.ATTRIB_IDEOGRAM_OVERLINE_DOUBLE
			p.buffer.Cursor.Attrib &^= buffer.ATTRIB_IDEOGRAM_STRESS_MARKING

		// Non default aixterm codes
		case 90, 91, 92, 93, 94, 95, 96, 97:
			p.buffer.Cursor.Color = n - 90
		case 100, 101, 102, 103, 104, 105, 106, 107:
			p.buffer.Cursor.Background = n - 100

		default: // Fallthrough
			log.Printf("unsupported SGR %d\n", n)
		}
	}

	return
}
