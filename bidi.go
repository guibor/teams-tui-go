package main

import (
	"strings"
	"unicode/utf8"

	"github.com/SCKelemen/unicode/uax9"
	"github.com/rivo/uniseg"
	"golang.org/x/text/unicode/bidi"
)

const (
	ansiEscape       = '\x1b'
	ansiHyperlinkEnd = "\x1b]8;;\x1b\\"
	ansiReset        = "\x1b[0m"
)

type ansiTextState struct {
	sgr  string
	link string
}

type ansiStyledRune struct {
	r     rune
	style ansiTextState
}

// bidiVisualLine converts logical Unicode text to the visual cell order needed
// by terminals without native bidirectional rendering. ANSI styles and OSC 8
// hyperlinks follow their original grapheme clusters through the reordering.
func bidiVisualLine(line string) (string, bool) {
	tokens := parseANSIStyledRunes(line)
	if len(tokens) == 0 {
		return line, false
	}

	runes := make([]rune, len(tokens))
	hasRTL := false
	for i, token := range tokens {
		runes[i] = token.r
		class := uax9.GetBidiClass(token.r)
		if class == uax9.ClassR || class == uax9.ClassAL {
			hasRTL = true
		}
	}
	if !hasRTL {
		return line, false
	}

	plain := string(runes)
	direction := uax9.GetParagraphDirection(plain)
	paragraphLevel := 0
	if direction == uax9.DirectionRTL {
		paragraphLevel = 1
	}

	classes := make([]uax9.BidiClass, len(runes))
	for i, r := range runes {
		classes[i] = uax9.GetBidiClass(r)
	}
	levels := uax9.ComputeLevels(classes, paragraphLevel)
	visualOrder := bidiVisualOrder(levels, paragraphLevel)
	clusters, clusterForRune := graphemeClusters(plain, len(runes))

	var out strings.Builder
	currentStyle := ansiTextState{}
	emittedCluster := make([]bool, len(clusters))
	for _, logicalIndex := range visualOrder {
		if logicalIndex < 0 || logicalIndex >= len(tokens) || levels[logicalIndex] < 0 {
			continue
		}
		clusterIndex := clusterForRune[logicalIndex]
		if emittedCluster[clusterIndex] {
			continue
		}
		emittedCluster[clusterIndex] = true

		for _, runeIndex := range clusters[clusterIndex] {
			if levels[runeIndex] < 0 {
				continue
			}
			token := tokens[runeIndex]
			writeANSIStateTransition(&out, currentStyle, token.style)
			currentStyle = token.style
			out.WriteRune(mirrorBidiRune(token.r, levels[runeIndex]))
		}
	}
	writeANSIStateTransition(&out, currentStyle, ansiTextState{})

	return out.String(), direction == uax9.DirectionRTL
}

func bidiVisualLines(lines []string) []string {
	visual := make([]string, len(lines))
	for i, line := range lines {
		visual[i], _ = bidiVisualLine(line)
	}
	return visual
}

// bidiVisualText applies terminal bidi ordering to every rendered line while
// preserving the logical source string held by the input component.
func bidiVisualText(text string) string {
	return strings.Join(bidiVisualLines(strings.Split(text, "\n")), "\n")
}

func bidiVisualOrder(levels []int, paragraphLevel int) []int {
	indices := make([]int, len(levels))
	maxLevel := paragraphLevel
	for i, level := range levels {
		indices[i] = i
		if level > maxLevel {
			maxLevel = level
		}
	}

	for level := maxLevel; level >= 1; level-- {
		for i := 0; i < len(levels); {
			if levels[i] >= 0 && levels[i] < level {
				i++
				continue
			}
			if levels[i] < 0 {
				i++
				continue
			}

			start := i
			i++
			for i < len(levels) && (levels[i] < 0 || levels[i] >= level) {
				i++
			}
			for left, right := start, i-1; left < right; left, right = left+1, right-1 {
				indices[left], indices[right] = indices[right], indices[left]
			}
		}
	}
	return indices
}

func graphemeClusters(text string, runeCount int) ([][]int, []int) {
	clusters := make([][]int, 0, runeCount)
	clusterForRune := make([]int, runeCount)
	runeIndex := 0
	iterator := uniseg.NewGraphemes(text)
	for iterator.Next() {
		clusterRunes := iterator.Runes()
		indices := make([]int, 0, len(clusterRunes))
		clusterIndex := len(clusters)
		for range clusterRunes {
			indices = append(indices, runeIndex)
			clusterForRune[runeIndex] = clusterIndex
			runeIndex++
		}
		clusters = append(clusters, indices)
	}
	return clusters, clusterForRune
}

func mirrorBidiRune(r rune, level int) rune {
	if level%2 == 0 {
		return r
	}
	properties, _ := bidi.LookupRune(r)
	if !properties.IsBracket() {
		return r
	}
	mirrored := []rune(bidi.ReverseString(string(r)))
	if len(mirrored) == 1 {
		return mirrored[0]
	}
	return r
}

func parseANSIStyledRunes(text string) []ansiStyledRune {
	tokens := make([]ansiStyledRune, 0, utf8.RuneCountInString(text))
	state := ansiTextState{}

	for i := 0; i < len(text); {
		if text[i] != ansiEscape {
			r, size := utf8.DecodeRuneInString(text[i:])
			tokens = append(tokens, ansiStyledRune{r: r, style: state})
			i += size
			continue
		}

		if i+1 >= len(text) {
			i++
			continue
		}
		switch text[i+1] {
		case '[':
			end := i + 2
			for end < len(text) && (text[end] < 0x40 || text[end] > 0x7e) {
				end++
			}
			if end >= len(text) {
				i = len(text)
				continue
			}
			end++
			sequence := text[i:end]
			if sequence[len(sequence)-1] == 'm' {
				if sgrHasFullReset(sequence) {
					state.sgr = sequence
					if sequence == ansiReset || sequence == "\x1b[m" {
						state.sgr = ""
					}
				} else {
					state.sgr += sequence
				}
			}
			i = end

		case ']':
			payloadStart := i + 2
			payloadEnd := payloadStart
			end := payloadStart
			for end < len(text) {
				if text[end] == '\a' {
					payloadEnd = end
					end++
					break
				}
				if text[end] == ansiEscape && end+1 < len(text) && text[end+1] == '\\' {
					payloadEnd = end
					end += 2
					break
				}
				end++
			}
			if payloadEnd == payloadStart && end >= len(text) {
				i = len(text)
				continue
			}
			payload := text[payloadStart:payloadEnd]
			if strings.HasPrefix(payload, "8;") {
				rest := payload[2:]
				if separator := strings.IndexByte(rest, ';'); separator >= 0 {
					if rest[separator+1:] == "" {
						state.link = ""
					} else {
						state.link = text[i:end]
					}
				}
			}
			i = end

		default:
			i += 2
		}
	}

	return tokens
}

func sgrHasFullReset(sequence string) bool {
	parameters := strings.TrimSuffix(strings.TrimPrefix(sequence, "\x1b["), "m")
	if parameters == "" {
		return true
	}
	for _, parameter := range strings.Split(parameters, ";") {
		if parameter == "0" {
			return true
		}
	}
	return false
}

func writeANSIStateTransition(out *strings.Builder, from, to ansiTextState) {
	if from.link != to.link && from.link != "" {
		out.WriteString(ansiHyperlinkEnd)
	}
	if from.sgr != to.sgr {
		if from.sgr != "" {
			out.WriteString(ansiReset)
		}
		if to.sgr != "" {
			out.WriteString(to.sgr)
		}
	}
	if from.link != to.link && to.link != "" {
		out.WriteString(to.link)
	}
}
