package parser

import (
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"

	"github.com/tehmaze-labs/go-piece/buffer"
	"github.com/tehmaze-labs/go-piece/calc"
	"github.com/tehmaze-labs/go-piece/color"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

const (
	ANSI_TABSTOP = 8
)

// ECMA-48 specified Final Bytes of control sequences without intermediate bytes
const (
	ANSI_ICH       = iota + 0x40 // '@', Insert Character
	ANSI_CUU                     // 'A', Cursor Up
	ANSI_CUD                     // 'B', Cursor Down
	ANSI_CUF                     // 'C', Cursor Right
	ANSI_CUB                     // 'D', Cursor Left
	ANSI_CNL                     // 'E', Cursor Next Line
	ANSI_CPL                     // 'F', Cursor Preceiding Line
	ANSI_CHA                     // 'G', Cursor Character Absolute
	ANSI_CUP                     // 'H', Cursor Position
	ANSI_CHT                     // 'I', Cursor Forward Tabulation
	ANSI_ED                      // 'J', Erase in Page
	ANSI_EL                      // 'K', Erase in Line
	ANSI_IL                      // 'L', Insert Line
	ANSI_DL                      // 'M', Delete Line
	ANSI_EF                      // 'N', Erase in Field
	ANSI_EA                      // 'O', Erase in Area
	ANSI_DCH                     // 'P', Delete Character
	ANSI_SSE                     // 'Q, ???
	ANSI_CPR                     // 'R', Active Position Report
	ANSI_SU                      // 'S', Scroll Up
	ANSI_SD                      // 'T', Scroll Down
	ANSI_NP                      // 'U', Next Page
	ANSI_PP                      // 'V', Preceding Page
	ANSI_CTC                     // 'W', Cursor Tabulation Control
	ANSI_ECH                     // 'X', Erase Character
	ANSI_CVT                     // 'Y', Cursor Line Tabulation
	ANSI_CBT                     // 'Z', Cursor Backward Tabulation
	ANSI_SRS                     // '[', Start Reversed String
	ANSI_PTX                     // '\', Paralell Texts
	ANSI_SDS                     // ']', Start Directed String
	ANSI_SIMD                    // '^', Select Implicit Movement Direction
	ANSI_UNDEFINED               // ' ', Unspecified
	ANSI_HPA                     // '`', Character Position Absolute
	ANSI_HPR                     // 'a', Character Position Forward
	ANSI_REP                     // 'b', Repeat
	ANSI_DA                      // 'c', Device Attributes
	ANSI_VPA                     // 'd', Line Position Absolute
	ANSI_VPR                     // 'e', Line Position Forward
	ANSI_HVP                     // 'f', Character and Line Position
	ANSI_TBC                     // 'g', Tabulation Clear
	ANSI_SM                      // 'h', Set Mode
	ANSI_MC                      // 'i', Media Copy
	ANSI_HPB                     // 'j', Character Position Absolute
	ANSI_VPB                     // 'k', Line Position Backward
	ANSI_RM                      // 'l', Reset Mode
	ANSI_SGR                     // 'm', Select Graphic Rendition
	ANSI_DSR                     // 'n', Device Status Report
	ANSI_DAQ                     // 'o', Define Area Qualification
)

type ansiOp func(seq *ANSISequence) error

type ANSI struct {
	Palette   color.Palette
	buffer    *buffer.Buffer
	opcode    map[byte]ansiOp
	transform transform.Transformer
}

func NewANSI(w, h int) *ANSI {
	p := &ANSI{
		Palette:   color.VGAPalette,
		buffer:    buffer.New(w, h),
		transform: charmap.CodePage437.NewDecoder(),
	}
	p.opcode = map[byte]ansiOp{
		ANSI_CHA: p.parseCHA,
		ANSI_CNL: p.parseCNL,
		ANSI_CPL: p.parseCHA,
		ANSI_CUB: p.parseCUB,
		ANSI_CUD: p.parseCUD,
		ANSI_CUF: p.parseCUF,
		ANSI_CUP: p.parseCUP,
		ANSI_CUU: p.parseCUU,
		ANSI_ED:  p.parseED,
		ANSI_EL:  p.parseEL,
		ANSI_IL:  p.parseIL,
		ANSI_HVP: p.parseCUP, // alias
		ANSI_SGR: p.parseSGR,
	}
	return p
}

func (p *ANSI) Parse(r io.Reader) (err error) {
	state := STATE_TEXT

	var seq = NewANSISequence()
	var buf = make([]byte, 1)
	var n int
	for state != STATE_EXIT {
		if n, err = r.Read(buf); n != 1 || err != nil {
			state = STATE_EXIT
			continue
		}

		ch := buf[0]

		switch state {
		case STATE_TEXT:
			switch ch {
			case SUB: // End Of File
				state = STATE_EXIT
			case ESC:
				state = STATE_ANSI_WAIT_BRACE
			case NL:
				p.buffer.Cursor.Y++
			case CR:
				p.buffer.Cursor.X = 0
			case TAB:
				c := (p.buffer.Cursor.X + 1) % ANSI_TABSTOP
				if c > 0 {
					c = ANSI_TABSTOP - c
					for i := 0; i < c; i++ {
						p.buffer.PutChar(' ')
					}
				}
			default:
				p.buffer.PutChar(ch)
			}

		case STATE_ANSI_WAIT_BRACE:
			if ch == '[' {
				state = STATE_ANSI_WAIT_LITERAL
			} else {
				p.buffer.PutChar(ESC)
				p.buffer.PutChar(ch)
			}

		case STATE_ANSI_WAIT_LITERAL:
			if ch == ';' {
				seq.Flush()
				break
			}

			if isAlpha(ch) {
				seq.Flush()
				//log.Printf("ANSI sequence <ESC>[%s%c (0x%02x)\n", seq, ch, ch)

				fn := p.opcode[ch]
				if fn == nil {
					log.Printf("Unsupported ANSI sequence <ESC>[%s%c (0x%02x)\n", seq, ch, ch)
				} else {
					if err = fn(seq); err != nil {
						log.Printf("Parser error: %v\n", err)
					}
				}

				seq.Reset()
				state = STATE_TEXT
				break
			} // if isAlpha(ch)
			seq.Buffer(ch)

		default:
			break
		}
	}

	w, h := p.buffer.SizeMax()
	log.Printf("screen at %d x %d\n", w+1, h+1)

	return nil
}

func (p *ANSI) Html() (s string) {
	s += "<!doctype html>\n"
	s += "<link rel=\"stylesheet\" href=\"cp437.css\">\n"
	s += "<style type=\"text/css\">\n"
	for i := 0; i < len(p.Palette); i++ {
		c := p.Palette[i].Hex()
		s += fmt.Sprintf(".f%02x{color:%s} ", i, c)
		s += fmt.Sprintf(".b%02x{background-color:%s} ", i, c)
		s += fmt.Sprintf(".u%02x{border-bottom:1px solid %s}", i, c)
		s += "\n"
	}
	s += `.i{font-variant:italics} .u{border-bottom:1px} .ud{border-bottom:3px dashed #000}`
	s += "</style>"

	s += fmt.Sprintf(`<pre><span class="b%02x f%02x">`,
		buffer.TILE_DEFAULT_BACKGROUND,
		buffer.TILE_DEFAULT_COLOR)

	w, h := p.buffer.SizeMax()
	var l *buffer.Tile

	for o, t := range p.buffer.Tiles {
		y, x := calc.DivMod(o, p.buffer.Width)
		if x >= w {
			continue
		}
		if y >= h {
			break
		}
		if x == 0 && y > 0 {
			s += "\n"
		}
		if t == nil {
			s += " "
		} else if t.Equal(l) {
			//s += string(t.Char)
			if isPrint(t.Char) {
				s += string(t.Char)
			} else {
				s += fmt.Sprintf(`&#x%02x;`, t.Char)
			}
		} else {
			f := t.Color
			b := t.Background
			c := []string{}

			if t.Attrib&buffer.ATTRIB_BOLD == buffer.ATTRIB_BOLD {
				f += 8
			}
			if t.Attrib&buffer.ATTRIB_BLINK == buffer.ATTRIB_BLINK {
				b += 8
			}
			if t.Attrib&buffer.ATTRIB_NEGATIVE == buffer.ATTRIB_NEGATIVE {
				f, b = b, f
			}
			c = append(c, fmt.Sprintf("b%02x", b))
			c = append(c, fmt.Sprintf("f%02x", f))
			if t.Attrib&buffer.ATTRIB_ITALICS > 0 {
				c = append(c, "i")
			}
			if t.Attrib&buffer.ATTRIB_UNDERLINE > 0 {
				c = append(c, fmt.Sprintf("u%02x", f))
			}
			if t.Attrib&buffer.ATTRIB_UNDERLINE_DOUBLE > 0 {
				c = append(c, "ud")
			}

			s += `</span>`
			s += fmt.Sprintf(`<span class="%s">`, strings.Join(c, " "))
			if isPrint(t.Char) {
				s += string(t.Char)
			} else {
				s += fmt.Sprintf(`&#x%02x;`, t.Char)
			}
		}

		l = t
	}

	s += `</span>`
	s += `</pre>`
	return
}

func (p *ANSI) String() (s string) {
	w, h := p.buffer.SizeMax()
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			o := (y * p.buffer.Width) + x
			t := p.buffer.Tile(o)
			if t == nil {
				s += " "
			} else {
				s += string(t.Char)
			}
		}
		s += "\n"
	}
	return
}

type ANSISequence struct {
	s []string
	b []byte
}

func NewANSISequence() *ANSISequence {
	return &ANSISequence{
		s: make([]string, 0),
		b: make([]byte, 0),
	}
}

func (s *ANSISequence) Buffer(b byte) {
	s.b = append(s.b, b)
}

func (s *ANSISequence) Flush() {
	s.s = append(s.s, string(s.b))
	s.b = make([]byte, 0)
}

func (s *ANSISequence) Int(n int) (i int) {
	if n < s.Len() {
		i, _ = strconv.Atoi(s.s[n])
	}
	return
}

func (s *ANSISequence) Ints() (i []int) {
	i = make([]int, 0)
	for _, j := range s.s {
		if n, err := strconv.Atoi(j); err == nil {
			i = append(i, n)
		}
	}
	return
}

func (s *ANSISequence) Len() int {
	return len(s.s)
}

func (s *ANSISequence) Reset() {
	s.s = make([]string, 0)
	s.b = make([]byte, 0)
}

func (s *ANSISequence) String() string {
	return strings.Join(s.s, ";")
}
