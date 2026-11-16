package userconfig

import (
	"regexp"
	"strings"
)

// formatText normalizes whitespace and Markdown, capitalizes, and ends the text with a dot.
func formatText(s string) string {
	return addDot(capitalize(formatMarkdown(normalizeSpace(s))))
}

var (
	reParagraphBreak = regexp.MustCompile(`\n\s*\n`)
	reSpaces         = regexp.MustCompile(`\s+`)
	reListItem       = regexp.MustCompile(`(?m)^\s*([-*+]|\d+\.)\s`)
)

// normalizeSpace unwraps hard-wrapped lines, keeping paragraphs, callouts, lists and code blocks.
func normalizeSpace(s string) string {
	s = strings.TrimSpace(s)
	if strings.Contains(s, "```") {
		return s
	}

	paragraphs := reParagraphBreak.Split(s, -1)
	for i, p := range paragraphs {
		p = strings.TrimSpace(p)
		if !HasCallout(p) && !reListItem.MatchString(p) {
			p = reSpaces.ReplaceAllString(p, " ")
		}
		paragraphs[i] = p
	}
	return strings.Join(paragraphs, "\n\n")
}

// reVerbatim matches what formatMarkdown keeps: code, and JSON objects without nested ones, e.g., {"key_name":"value"}.
var reVerbatim = regexp.MustCompile("(?s)```.*?```|`[^`]*`" + `|\{\s*(?:"|&quot;)[^{}]*\}`)

// symbolRun matches a run of symbol characters: no letters (\pL), digits (\pN),
// whitespace or quotes. It is spelled out rather than using \W so that "_",
// which \w counts as a word character, is treated as a symbol too. Excluding
// whitespace is what keeps the gap between two quoted values from being
// mistaken for a quoted symbol: a quoted token such as ".*" never contains
// whitespace, while every such gap does (e.g. the ": " in {"a": "b"}).
const symbolRun = `([^\pL\pN\s'"]+)`

// markdownFixes render plain text as intended in Markdown, in order.
var markdownFixes = []func(string) string{
	// Single-quoted tokens, e.g. 'aiven-'.
	replaceAll(`'([^'\s]+)'`, "`$1`"),
	// Double-quoted symbols, e.g. "/" or ".*", also as the &quot; entity that the OpenAPI spec ships.
	// The opening quote must not follow a value, or the gap in ["a","b"] would match.
	replaceAll(`(^|[\s(\[,])"`+symbolRun+`"`, "$1`$2`"),
	replaceAll(`(^|[\s(\[,])&quot;`+symbolRun+`&quot;`, "$1`$2`"),
	// Wildcards, e.g. public-* or a lone *, which Markdown parses as emphasis.
	// A word's * must follow "-" or "_", or it might close emphasis, e.g. **bold text**.
	replaceAll(`(^|\s)([\w-]*[-_]\*[\w*-]*)`, "$1`$2`"),
	// A * after a line's indentation is a list item.
	replaceAll(`(^|\S)([ \t]+)\*([ \t])`, "$1$2`*`$3"),
	// Placeholders, e.g. ${name} or <group ID>, which Markdown parses as HTML tags.
	codePlaceholders,
	outsideLinks(codeIdentifiers),
	outsideLinks(codeNumbers),
}

func replaceAll(expr, repl string) func(string) string {
	re := regexp.MustCompile(expr)
	return func(s string) string { return re.ReplaceAllString(s, repl) }
}

// formatMarkdown applies markdownFixes outside reVerbatim.
func formatMarkdown(s string) string {
	for _, fix := range markdownFixes {
		var b strings.Builder
		last := 0
		for _, m := range reVerbatim.FindAllStringIndex(s, -1) {
			b.WriteString(fix(s[last:m[0]]))
			b.WriteString(s[m[0]:m[1]])
			last = m[1]
		}
		b.WriteString(fix(s[last:]))
		s = b.String()
	}
	return s
}

// rePlaceholder matches a "${...}" placeholder, or a word with `<placeholder>` parts, e.g. `${<a>:<b c>}`,
// but not the brackets or quotes around it.
var rePlaceholder = regexp.MustCompile(`\$\{[^}\s]*\}|` +
	"[^\\s`()\\[\\]\"'&;]*<\\w[^<>]*>(?:[^\\s<>`]*<\\w[^<>]*>)*[^\\s`()\\[\\]\"'&;]*")

func codePlaceholders(s string) string {
	var b strings.Builder
	last := 0
	for _, m := range rePlaceholder.FindAllStringIndex(s, -1) {
		start := m[0]
		end := start + len(strings.TrimRight(s[start:m[1]], ".,:!?"))
		before, after := quoted(s, start, end)
		b.WriteString(s[last:before])
		b.WriteString("`" + s[start:end] + "`")
		last = after
	}
	b.WriteString(s[last:])
	return b.String()
}

// quoted returns the bounds of s[start:end] with the double quotes around it, if any.
func quoted(s string, start, end int) (int, int) {
	for _, q := range []string{`"`, "&quot;"} {
		if strings.HasSuffix(s[:start], q) && strings.HasPrefix(s[end:], q) {
			return start - len(q), end + len(q)
		}
	}
	return start, end
}

// glued reports whether c makes a match part of a longer token, e.g., "Max_size" or "path/to_file".
func glued(c byte, symbols string) bool {
	return 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || '0' <= c && c <= '9' || strings.IndexByte(symbols, c) >= 0
}

// reNumber matches a negative number, or a dotted number, e.g., a version, a decimal, or an IP address with a CIDR suffix.
var reNumber = regexp.MustCompile(`-\d+(?:\.\d+)*|\d+(?:\.\d+)+(?:/\d+)?`)

// codeNumbers wraps the numbers that stand alone in backticks,
// e.g., not "v1.2", "1.2.3-rc", "0.5%", "25.8+", "10.0.0.1:9092" or "2024-01-01".
func codeNumbers(s string) string {
	isSpace := func(i int) bool { return i == len(s) || strings.IndexByte(" \t\n", s[i]) >= 0 }

	var b strings.Builder
	last := 0
	for _, m := range reNumber.FindAllStringIndex(s, -1) {
		start, end := m[0], m[1]
		before, after := quoted(s, start, end)
		standsAlone := (before == 0 || strings.IndexByte(" \t\n([", s[before-1]) >= 0) &&
			(isSpace(after) || strings.IndexByte(")],;", s[after]) >= 0 ||
				strings.IndexByte(".:", s[after]) >= 0 && isSpace(after+1)) // e.g., "Default is 0.5." or "Kafka 4.0: ..."
		if !standsAlone {
			continue
		}
		b.WriteString(s[last:before])
		b.WriteString("`" + s[start:end] + "`")
		last = after
	}
	b.WriteString(s[last:])
	return b.String()
}

// reLink matches URLs and Markdown link targets.
var reLink = regexp.MustCompile(`https?://\S+|\]\([^)]*\)`)

// outsideLinks applies fix outside reLink.
func outsideLinks(fix func(string) string) func(string) string {
	return func(s string) string {
		var b strings.Builder
		last := 0
		for _, m := range reLink.FindAllStringIndex(s, -1) {
			b.WriteString(fix(s[last:m[0]]))
			b.WriteString(s[m[0]:m[1]])
			last = m[1]
		}
		b.WriteString(fix(s[last:]))
		return b.String()
	}
}

// reIdentifier matches a lowercase or uppercase word or dotted path,
// e.g., "pg_stat_monitor.pgsm_max_buckets" or "PERFORMANCE_SCHEMA.EVENT_WAITS".
var reIdentifier = regexp.MustCompile(`[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*|[A-Z][A-Z0-9_]*(?:\.[A-Z][A-Z0-9_]*)*`)

// reSettingValue matches the value after an identifier, e.g., "=worker" of "io_method=worker".
var reSettingValue = regexp.MustCompile(`^=[^\s,;)'"&]*`)

var reDottedPath = regexp.MustCompile(`^[a-z0-9]+(?:\.[a-z0-9]+)+$`)

// isDottedPath reports whether id is a lowercase dotted path, e.g., "min.insync.replicas" but not "e.g".
func isDottedPath(id string) bool {
	return reDottedPath.MatchString(id) && len(id) > len(strings.Split(id, "."))*2-1
}

// codeIdentifiers wraps the lowercase and uppercase words with "_" and the lowercase dotted paths
// in backticks, e.g., sas_token, AIVEN_TOKEN and log.retention.ms, replacing the quotes around them.
func codeIdentifiers(s string) string {
	var b strings.Builder
	last := 0
	for _, m := range reIdentifier.FindAllStringIndex(s, -1) {
		start, end := m[0], m[1]
		id := s[start:end]
		switch {
		case start < last, // In the value of a setting
			!strings.Contains(id, "_") && !isDottedPath(id),
			// Part of a longer token, e.g., "a-b_c", "my_user@host" or "Long.MAX_VALUE"
			start > 0 && glued(s[start-1], `_-/@.$\`),
			end < len(s) && glued(s[end], `-/@\`):
			continue
		}

		// A setting, e.g., "io_method=worker"
		if v := reSettingValue.FindString(s[end:]); v != "" {
			end += len(strings.TrimRight(v, "."))
		} else if strings.HasPrefix(s[end:], "()") {
			end += len("()") // A function, e.g., "GROUP_CONCAT()"
		}

		before, after := start, end
		for _, q := range []string{"'", `"`, "&quot;"} {
			if strings.HasSuffix(s[:start], q) && strings.HasPrefix(s[end:], q) {
				before, after = start-len(q), end+len(q)
				break
			}
		}
		// A JSON key or string value, e.g., {"key_name": "value_name"}
		isQuoted := after > end
		jsonKey := isQuoted && strings.HasPrefix(s[after:], ":")
		jsonValue := isQuoted && s[before] != '\'' && strings.HasSuffix(strings.TrimRight(s[:before], " "), ":")
		if jsonKey || jsonValue {
			continue
		}

		b.WriteString(s[last:before])
		b.WriteString("`" + s[start:end] + "`")
		last = after
	}
	b.WriteString(s[last:])
	return b.String()
}

// reLowerWord matches a leading lowercase word, but not a literal, e.g. "sha256" or "user_operations".
var reLowerWord = regexp.MustCompile(`^[a-z]+[\s,]`)

func capitalize(s string) string {
	if strings.HasPrefix(s, "fnmatch") || !reLowerWord.MatchString(s) {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// addDot ends a sentence with a single dot, keeping "!", "?" and a closing code fence.
func addDot(s string) string {
	s = strings.TrimRight(s, ". \t\n")
	for _, end := range []string{"!", "?", "```"} {
		if strings.HasSuffix(s, end) {
			return s
		}
	}
	return s + "."
}
