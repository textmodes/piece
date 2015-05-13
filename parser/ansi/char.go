package ansi

// ASCII escape characters
const (
	NUL = iota
	SOH
	STX
	ETX
	EOT
	ENQ
	ACK
	BEL
	BS
	TAB
	LF
	VT
	FF
	CR
	SO
	SI
	DLE
	DC1
	DC2
	DC3
	DC4
	NAK
	SYN
	ETB
	CAN
	EM
	SUB
	ESC
	FS
	GS
	RS
	US
	Space

	NL = LF
	NP = FF
)

func isAlpha(c byte) bool {
	return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func isPrint(c byte) bool {
	return c >= 0x20 && c < 0x7f
}
