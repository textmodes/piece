package main

import (
	"flag"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"git.maze.io/maze/go-piece/font"
	"git.maze.io/maze/go-piece/parser"
	"git.maze.io/maze/go-piece/parser/ansi"
	"git.maze.io/maze/go-piece/parser/binarytext"
	"git.maze.io/maze/go-piece/parser/irc"
	"git.maze.io/maze/go-piece/parser/xbin"
	sauce "git.maze.io/maze/go-sauce"
)

var supportedParser = [][]string{
	[]string{"ANSi/ASCII", "ansi", "ascii", "text"},
	[]string{"Binary text (raw VGA page)", "bin", "binarytext"},
	[]string{"IRC log with mIRC formatting", "irc", "mirc"},
	[]string{"eXtended Binary text", "xbin"},
}

func guessParser(filename, option string) parser.Parser {
	if option != "" {
		switch strings.ToLower(option) {
		case "help", "list":
			fmt.Fprintln(os.Stderr, "Supported parsers:")
			for _, p := range supportedParser {
				var (
					o = p[0]
					a = p[1:]
				)
				fmt.Fprintf(os.Stderr, "\n\t%s:\n", strings.Join(a, ", "))
				fmt.Fprintf(os.Stderr, "\t\t%s\n", o)
			}
			fmt.Fprintln(os.Stderr, "")
			return nil
		case "ansi", "ascii", "text":
			return ansi.New(80, 25)
		case "bin", "binarytext":
			return binarytext.New()
		case "irc", "mirc":
			return irc.New()
		case "xbin":
			return xbin.New()
		}
	}

	switch strings.ToLower(filepath.Ext(filename)) {
	case ".asc", ".ans", ".txt", ".diz", ".lit":
		return ansi.New(80, 25)

	case ".bin":
		return binarytext.New()

	case ".irc", ".log":
		return irc.New()

	case ".xb":
		return xbin.New()
	}

	return nil
}

func main() {
	formatFlag := flag.String("format", "html", "Output format")
	outputFlag := flag.String("output", "", "Output filename")
	parserFlag := flag.String("parser", "", "Parser (default: autodetect)")
	fontFlag := flag.String("font", "", "Font override")
	fontSizeFlag := flag.String("font-size", "", "Font size override")
	defaultFontFlag := flag.String("default-font", "cp437", "Default font")
	defaultFontSizeFlag := flag.String("default-font-size", "8x16", "Default font size")
	flag.Parse()

	if len(flag.Args()) != 1 {
		fmt.Fprintln(os.Stderr, "Error: missing filename")
		flag.Usage()
		os.Exit(1)
	}

	var filename = flag.Args()[0]
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	var o io.Writer
	switch *outputFlag {
	case "", "-", "stdout", "/dev/stdout":
		o = os.Stdout

	case "stderr", "/dev/stderr":
		o = os.Stderr

	default:
		var of *os.File
		if of, err = os.Create(*outputFlag); err != nil {
			log.Fatalln("%s: error creating %s: %v\n", filename, *outputFlag, err)
		}
		defer of.Close()
		o = of
	}

	i, err := f.Stat()
	if err != nil {
		panic(err)
	}

	var p parser.Parser
	w, h := 80, 25
	r := io.NewSectionReader(f, 0, i.Size())
	s, err := sauce.Parse(r)
	if err != nil && err != sauce.ErrNoRecord {
		log.Printf("%s: failed to parse SAUCE: %v\n", filename, err)
	}
	if s == nil || *parserFlag != "" {
		p = guessParser(filename, *parserFlag)

	} else {
		switch s.DataType {
		case sauce.DataTypeCharacter:
			switch s.FileType {
			case 0, 1:
				w = int(s.TInfo[0])
				p = ansi.New(w, h)
			}

		case sauce.DataTypeBinaryText:
			p = binarytext.New()

		case sauce.DataTypeXBIN:
			p = xbin.New()

		default:
			log.Printf("%s: unsupported data type %s (%02x)\n", filename, s.DataTypeString(), s.DataType)
		}
	}

	if p == nil {
		log.Fatalf("%s: no suitable parser found\n", filename)
	}

	if _, err = r.Seek(0, 0); err != nil {
		log.Fatalf("%s: rewind failed: %v\n", filename, err)
	}
	if err = p.Parse(r); err != nil {
		log.Fatalf("%s: parse failed: %v\n", filename, err)
	}

	switch *formatFlag {
	case "html":
		var html string
		if html, err = p.HTML(true); err != nil {
			log.Fatalf("%s: render failed: %v\n", filename, err)
		}
		fmt.Fprint(o, html)

	case "image", "gif", "jpg", "jpeg", "png":
		var pieceFont = p.Font()
		if pieceFont == nil {
			if *fontFlag == "" {
				*fontFlag = *defaultFontFlag
			}
			if *fontSizeFlag == "" {
				*fontSizeFlag = *defaultFontSizeFlag
			}

			var fontSize image.Point
			if fontSize, err = font.ParseSize(*fontSizeFlag); err != nil {
				log.Fatalf("%s: %v\n", filename, err)
			}
			pieceFont = font.Get(*fontFlag, fontSize)
		}

		var i image.Image
		if i, err = p.Image(pieceFont); err != nil || i == nil {
			log.Fatalln("%s: render failed: %v\n", filename, err)
		}

		switch *formatFlag {
		case "gif":
			err = gif.Encode(o, i, nil)

		case "jpeg", "jpg":
			err = jpeg.Encode(o, i, nil)

		case "image", "png":
			err = png.Encode(o, i)
		}
		if err != nil {
			log.Fatalln("%s: encode failed: %v\n", filename, err)
		}

	case "text":
		fmt.Fprint(o, p.String())

	default:
		log.Fatalf("Unknown format %q\n", *formatFlag)
	}
}
