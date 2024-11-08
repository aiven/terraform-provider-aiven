package main

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	draftVersion         = "MAJOR.MINOR.PATCH"
	draftDate            = "YYYY-MM-DD"
	defaultBullet        = "- "
	defaultLineMaxLength = 120 // Line soft wrap settings
)

type changelogItem struct {
	Date, Version, Content string
}

var (
	reVersion       = regexp.MustCompile(`\w+\.\w+\.\w+`)
	reDate          = regexp.MustCompile(`\w{4}-\w{2}-\w{2}`)
	reSplitEntries  = regexp.MustCompile(`(?m)^(\b[^a-z]| *- +)`) // A line that begins with "-" or a non-letter
	reBulletLevel   = regexp.MustCompile(`^ *- +`)
	reSpaces        = regexp.MustCompile(`\s+`)
	reTrailingSpace = regexp.MustCompile(`\s+$`)
)

// updateChangelog updates the changelog with the given addLines
// Soft-wraps lines to the given lineLength
// When reformat is true, reformats the whole given content
func updateChangelog(content string, lineLength int, reformat bool, addLines ...string) (string, error) {
	if addLines == nil && !reformat {
		return content, nil
	}

	lines := strings.Split(reTrailingSpace.ReplaceAllString(content, ""), "\n")
	items, start, end := parseItems(lines)
	addText := strings.Join(addLines, "\n")

	if len(items) != 0 && items[0].Version == draftVersion {
		// Appends to the current draft
		items[0].Content = fmt.Sprintf("%s\n%s", items[0].Content, addText)
	} else {
		// The First item is not the draft, so we need to add a new item
		items = append(items, &changelogItem{
			Version: draftVersion,
			Date:    draftDate,
			Content: content,
		})
	}

	result := lines[:start]
	for i, v := range items {
		c := strings.TrimSpace(v.Content)
		if i == 0 || reformat {
			c = formatContent(c, lineLength)
		}
		header := fmt.Sprintf("## [%s] - %s", v.Version, v.Date)
		result = append(result, header, "", c, "") // Empty lines for readability and formatting
	}

	result = append(result, lines[end+1:]...) // Adds the rest of the file
	return strings.Join(result, "\n"), nil
}

func parseItems(lines []string) ([]*changelogItem, int, int) {
	start := max(0, len(lines)-1)
	end := start
	var item *changelogItem
	items := make([]*changelogItem, 0)
	for i, line := range lines {
		if strings.HasPrefix(line, "##") {
			if item == nil {
				start = i
			}

			item = &changelogItem{
				Date:    reDate.FindString(line),
				Version: reVersion.FindString(line),
			}

			items = append(items, item)
			continue
		}

		if line != "" && item != nil {
			item.Content = item.Content + strings.TrimSuffix(line, " ") + "\n"
			end = i
		}
	}
	return items, start, end
}

func formatContent(content string, lineLength int) string {

	// Golang doesn't support regexp "lookarounds", so we need to split the content,
	// and then join it to keep what we otherwise would be just ignored by negative lookbehind
	seps := reSplitEntries.FindAllStringSubmatchIndex(content, -1)
	chunks := reSplitEntries.Split(content, -1)
	list := make([]string, 0)
	seen := make(map[string]bool)
	for i, v := range seps {
		// This is the separator between the entries
		sep := content[v[0]:v[1]]

		// Joins with the separator in case it has "negative lookbehind" part
		text := strings.TrimRight(sep+chunks[i+1], "\n ")

		// Looks for the bullet
		bullet := reBulletLevel.FindString(text)
		if bullet != "" {
			// When found, separates the text
			text = strings.SplitN(text, bullet, 2)[1]
		} else {
			// Otherwise, uses the default bullet
			bullet = defaultBullet
		}

		// Removes original spaces and newlines
		point := addBullet(bullet, softWrap(reSpaces.ReplaceAllString(text, " "), lineLength))

		// Removes duplicates
		if !seen[point] {
			seen[point] = true
			list = append(list, point)
		}
	}

	return strings.Join(list, "\n")
}

var reShortWords = regexp.MustCompile(`(\b.{1,3}\b) +`)

// softWrap wraps text to a given size
// Keeps prepositions and articles together with the next word for better readability
func softWrap(text string, size int) []string {
	text = reShortWords.ReplaceAllString(text, "$1⍽")

	j := 0
	result := make([]string, 1)
	for i, w := range strings.Split(text, " ") {
		w = strings.ReplaceAll(w, "⍽", " ")
		switch {
		case i == 0:
			result[j] += w
		case len(result[j])+len(w) < size:
			result[j] += " " + w
		default:
			result = append(result, w) // nolint: makezero // By some reason linter doesn't understand it has length 1
			j++
		}
	}
	return result
}

// addBullet add the given bullet to the beginning of the first line and indents the rest
func addBullet(bullet string, lines []string) string {
	prefix := strings.Repeat(" ", len(bullet))
	for i, v := range lines {
		if i == 0 {
			lines[i] = bullet + v
		} else {
			lines[i] = prefix + v
		}
	}
	return strings.Join(lines, "\n")
}
