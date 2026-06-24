package main

import "strings"

// plainText converts the small subset of Markdown that the description builder
// emits into plain text for the Terraform registry front matter `description:`
// field. Only that metadata field must be plain text; the visible doc page keeps
// the full Markdown.
//
// This is a purpose-built converter (no Markdown library, no AST) that scans the
// text line by line and rewrites the inline constructs that actually appear in
// generated descriptions:
//   - inline code spans keep their literal contents, markers removed
//   - emphasis (*, **) and strong/strikethrough markers are removed, text kept
//   - links render as "text URL" for absolute targets and "text" for fragment or
//     path-relative targets; images are dropped entirely
//   - blank lines are collapsed so paragraphs are joined by a single newline
func plainText(markdown string) string {
	lines := strings.Split(markdown, "\n")
	kept := make([]string, 0, len(lines))
	for _, line := range lines {
		stripped := strings.TrimRight(stripInlineMarkdown(line), " \t")
		if strings.TrimSpace(stripped) == "" {
			continue
		}
		kept = append(kept, stripped)
	}
	return strings.Join(kept, "\n")
}

// stripInlineMarkdown rewrites a single line, removing inline Markdown markers.
func stripInlineMarkdown(line string) string {
	src := []rune(line)
	var out strings.Builder
	for i := 0; i < len(src); {
		switch src[i] {
		case '`':
			content, next, ok := readCodeSpan(src, i)
			if !ok {
				out.WriteRune('`')
				i++
				continue
			}
			out.WriteString(content)
			i = next
		case '!':
			if text, _, next, ok := readLink(src, i+1); ok {
				_ = text // images contribute no text
				i = next
				continue
			}
			out.WriteRune('!')
			i++
		case '[':
			text, dest, next, ok := readLink(src, i)
			if !ok {
				out.WriteRune('[')
				i++
				continue
			}
			out.WriteString(stripInlineMarkdown(text))
			if dest != "" && !isRelativeTarget(dest) {
				out.WriteByte(' ')
				out.WriteString(dest)
			}
			i = next
		case '*':
			for i < len(src) && src[i] == '*' {
				i++
			}
		case '~':
			// "~~" is strikethrough; a lone "~" (e.g. the "~>" callout) is kept.
			if i+1 < len(src) && src[i+1] == '~' {
				i += 2
				continue
			}
			out.WriteRune('~')
			i++
		case '_':
			// "__" is strong; a lone "_" is kept so identifiers and paths survive.
			if i+1 < len(src) && src[i+1] == '_' {
				i += 2
				continue
			}
			out.WriteRune('_')
			i++
		default:
			out.WriteRune(src[i])
			i++
		}
	}
	return out.String()
}

// readCodeSpan reads a backtick code span starting at src[start]. It matches a
// closing run of the same number of backticks and returns the literal contents.
func readCodeSpan(src []rune, start int) (content string, next int, ok bool) {
	fence := 0
	for start+fence < len(src) && src[start+fence] == '`' {
		fence++
	}
	contentStart := start + fence
	for i := contentStart; i < len(src); i++ {
		if src[i] != '`' {
			continue
		}
		run := 0
		for i+run < len(src) && src[i+run] == '`' {
			run++
		}
		if run == fence {
			return string(src[contentStart:i]), i + run, true
		}
		i += run - 1
	}
	return "", 0, false
}

// readLink parses "[text](dest)" starting at src[start] == '['. Any link title
// after the destination is discarded. It returns the text, destination, the
// index just past the closing ')', and whether parsing succeeded.
func readLink(src []rune, start int) (text, dest string, next int, ok bool) {
	if start >= len(src) || src[start] != '[' {
		return "", "", 0, false
	}
	i := start + 1
	textStart := i
	depth := 1
	for ; i < len(src) && depth > 0; i++ {
		switch src[i] {
		case '[':
			depth++
		case ']':
			depth--
			if depth == 0 {
				goto closeText
			}
		}
	}
	return "", "", 0, false

closeText:
	text = string(src[textStart:i])
	i++ // past ']'
	if i >= len(src) || src[i] != '(' {
		return "", "", 0, false
	}
	i++ // past '('
	destStart := i
	for i < len(src) && src[i] != ')' && src[i] != ' ' {
		i++
	}
	dest = string(src[destStart:i])
	for i < len(src) && src[i] != ')' {
		i++
	}
	if i >= len(src) {
		return "", "", 0, false
	}
	return text, dest, i + 1, true
}

// isRelativeTarget reports whether a link destination points within the page or
// site (fragment or path-relative), in which case only the link text is kept.
func isRelativeTarget(dest string) bool {
	switch dest[0] {
	case '#':
		return true
	case '/':
		return len(dest) == 1 || dest[1] != '/'
	default:
		return false
	}
}
