package font

var fontAlias = map[string]string{
	"default": "cp437",

	// SAUCE v00.5 compatible names
	"Amiga MicroKnight":  "microknight",
	"Amiga MicroKnight+": "microknightplus",
	"Amiga mOsOul":       "mo-soul",
	"Amiga P0T-NOoDLE":   "p0t-noodle",
	"Amiga Topaz 1":      "topaz-a500",
	"Amiga Topaz 1+":     "topazplus-a500",
	"Amiga Topaz 2":      "topaz-a1200",
	"Amiga Topaz 2+":     "topazplus-a1200",
	"Atari ATASCII":      "atascii-international",
	"IBM VGA":            "cp437",
	"IBM VGA50":          "cp437",
}

var codePages = []string{
	"437",
	"720",
	"737",
	"775",
	"819",
	"850",
	"852",
	"855",
	"857",
	"858",
	"860",
	"861",
	"862", // Hebrew
	"863", // French Canada
	"864", // Arabic
	"865", // Nordic
	"866", // Cyrillic
	"869", // Greek 2
	"872", // Cyrillic
}

func init() {
	for _, codePage := range codePages {
		fontAlias["IBM VGA "+codePage] = "cp" + codePage
		fontAlias["IBM VGA25G "+codePage] = "cp" + codePage
		fontAlias["IBM VGA50 "+codePage] = "cp" + codePage
		fontAlias["IBM EGA "+codePage] = "cp" + codePage
		fontAlias["IBM EGA43 "+codePage] = "cp" + codePage
	}
}
