package attribute

const (
	None                    = 0 << iota // No attributes
	Bold                                // bold or increased intensity
	Faint                               // faint, decreased intensity or second colour
	Italics                             // italicized
	Underline                           // underlined
	Blink                               // blinking
	Negative                            // negative image
	Conceal                             // concealed characters
	CrossedOut                          // crossed-out (characters still legible but marked as to be deleted)
	Gothic                              // Fraktur (gothic)
	DoubleUnderline                     // doubly underlined
	Frame                               // framed
	Encircle                            // encircled
	Overline                            // overlined
	IdeogramUnderline                   // ideogram underline or right side line
	IdeogramDoubleUnderline             // ideogram double underline or double line on the right side
	IdeogramOverline                    // ideogram overline or left side line
	IdeogramDoubleOverline              // ideogram double overline or double line on the left side
	IdeogramStressMarking               // ideogram stress marking
)
