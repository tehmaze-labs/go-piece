package parser

const (
	STATE_EXIT = iota
	STATE_TEXT
	STATE_ANSI_WAIT_BRACE
	STATE_ANSI_WAIT_LITERAL
)
