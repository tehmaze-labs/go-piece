package buffer

import (
	"errors"

	"github.com/tehmaze-labs/go-piece/calc"
)

var (
	errOutOfBounds = errors.New("Out of bounds")
)

type Buffer struct {
	Width, Height       int
	Cursor              *Cursor
	Tiles               []*Tile
	maxWidth, maxHeight int
}

// New creates a new buffer of w x h Tiles. The maximum buffer width is set to
// the supplied width w.
func New(w, h int) *Buffer {
	b := &Buffer{
		Width:  w,
		Height: h,
		Cursor: NewCursor(0, 0),
		Tiles:  make([]*Tile, w*h),
	}
	b.Resize(w, h)
	return b
}

// Clear removes all Tiles from the screen
func (b *Buffer) Clear() {
	for o := range b.Tiles {
		b.Tiles[o] = nil
	}
}

// ClearAt clears a tile at offset o
func (b *Buffer) ClearAt(o int) {
	if o < len(b.Tiles) {
		b.Tiles[o] = nil
	}
}

// ClearFrom clears all Tiles from offset o
func (b *Buffer) ClearFrom(o int) {
	for l := len(b.Tiles); o < l; o++ {
		b.Tiles[o] = nil
	}
}

// ClearTo clears all Tiles up until offset o
func (b *Buffer) ClearTo(o int) {
	for o = calc.MinInt(o, len(b.Tiles)); o >= 0; o-- {
		b.Tiles[o] = nil
	}
}

// Insert inserts n Tiles at offset o.
func (b *Buffer) Insert(o, n int) {
	p := make([]*Tile, n)
	b.Tiles = append(b.Tiles[:o], append(p, b.Tiles[o:]...)...)
}

// Expand buffer to fit offset o.
func (b *Buffer) Expand(o int) *Buffer {
	l := len(b.Tiles)
	if l <= o {
		b.Tiles = append(b.Tiles, make([]*Tile, o-l+1)...)
	}
	return b
}

// Len returns the number of possible Tiles (total offset)
func (b *Buffer) Len() int {
	return b.Width * b.Height
}

// Normalize fits the cursor within the canvas.
func (b *Buffer) Normalize() *Buffer {
	w, h := b.Size()
	b.Cursor.Normalize(w, h)
	return b
}

// Resize to at least a w x h canvas. If w exceeds the maximum buffer width an
// error is thrown.
func (b *Buffer) Resize(w, h int) (*Buffer, error) {
	return nil, errors.New("Not implemented")
}

// Size calculates the allocated buffer size.
func (b *Buffer) Size() (w int, h int) {
	var l int
	l = b.Len()
	w = b.Width
	if w > 0 {
		h = 1 + ((l - 1) / w)
	}
	return
}

// SizeMax returns the actual used buffer size.
func (b *Buffer) SizeMax() (int, int) {
	return b.maxWidth, b.maxHeight
}

// Tile at offset o, will allocate a new Tile if it doesn't exist at the
// requested offset.
func (b *Buffer) Tile(o int) *Tile {
	if o >= len(b.Tiles) {
		return nil
	}
	if b.Tiles[o] == nil {
		b.Tiles[o] = NewTile()
	}
	return b.Tiles[o]
}

// PutChar writes a character to the buffer at the current cursor location and
// advances the cursor position.
func (b *Buffer) PutChar(c byte) error {
	b.Cursor.Char = c
	o := b.Cursor.Offset(b.Width)
	t := b.Expand(o).Tile(o)
	t.Update(&b.Cursor.Tile)
	b.Cursor.X++
	b.Cursor.NormalizeAndWrap(b.Width)
	b.maxWidth = calc.MaxInt(b.maxWidth, b.Cursor.X)
	b.maxHeight = calc.MaxInt(b.maxHeight, b.Cursor.Y)
	return nil
}
