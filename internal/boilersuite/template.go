/*
Copyright 2023 The cert-manager Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package boilersuite

import (
	"fmt"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/hexops/gotextdiff/span"
)

// Pre-processed template ready for use during validation.
type Template struct {
	// Text of the template, sanity-checked and with the <<AUTHOR>> marker replaced
	text string
	// Optional parsing step, for example, to skip go build constraints
	skipHeaderFunc func(string) int
}

// Create a new boilerplate template using the given content and configuration
func NewTemplate(content string, ext string, expectedAuthor string) (Template, error) {
	// Sanity-check
	if !strings.Contains(content, CopyrightMarker) {
		return Template{}, fmt.Errorf("couldn't find replacement marker %q", CopyrightMarker)
	}

	if !strings.Contains(content, AuthorMarker) {
		return Template{}, fmt.Errorf("couldn't find replacement marker %q", AuthorMarker)
	}

	if strings.Contains(content, "\r") {
		return Template{}, fmt.Errorf("has Windows style line endings. Unix style are required")
	}

	// Edit content
	text := strings.ReplaceAll(content, AuthorMarker, expectedAuthor)
	text = strings.TrimSpace(text) + "\n"

	// Find skipHeaderFunc
	var skipHeaderFunc func(string) int
	switch ext {
	case "go":
		skipHeaderFunc = skipHeaderGoFile
	case "sh", "bash", "py":
		skipHeaderFunc = skipHeaderShebang
	}

	return Template{
		text:           text,
		skipHeaderFunc: skipHeaderFunc,
	}, nil
}

// Validate checks the given file path against the template
func (t Template) Validate(path string, patch bool) error {
	// Read file and check
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read: %w", err)
	}
	return t.validateContent(string(content), path, patch)
}

func (t Template) validateContent(content string, path string, patch bool) error {
	if SkipFileRegex.MatchString(content) || GeneratedRegex.MatchString(content) {
		return nil
	}

	// Find existing boilerplate year and location
	head, boilOrig, foot, year := t.analyzeFile(content)
	boilExpect := strings.ReplaceAll(t.text, YearMarker, year)

	// Build the expected content, using the same newline type as the original, and ensuring exactly one empty line
	// before/after the boilerplate
	nl := "\n"
	if strings.Contains(content, "\r\n") {
		boilExpect = strings.ReplaceAll(boilExpect, "\n", "\r\n")
		nl = "\r\n"
	}
	head = strings.TrimSpace(head)
	if head != "" {
		head += nl + nl
	}
	if foot != "" {
		foot = nl + strings.TrimLeftFunc(foot, unicode.IsSpace)
	}
	expect := head + boilExpect + foot

	// Return error and patch if we don't have what we want
	if content != expect {
		reason := "incorrect boilerplate"
		if boilOrig == "" {
			reason = "missing boilerplate"
		}
		if patch {
			edits := myers.ComputeEdits(span.URIFromPath(path), content, expect)
			return fmt.Errorf("%s\n%s", reason, gotextdiff.ToUnified(path, "expected", content, edits))
		} else {
			return fmt.Errorf("%s", reason)
		}
	}

	return nil
}

// Split the input into header/boilerplate/footer parts, and find the copyright year.
// The boilerplate part may be empty, and in this case the copyright year is generated.
// The header might be a shebang, golang build constraints, etc (see LoadTemplates).
func (t Template) analyzeFile(content string) (head string, boil string, foot string, year string) {
	start, stop, year := findExistingBoilerplate(content)
	if start == -1 {
		if t.skipHeaderFunc != nil {
			start = t.skipHeaderFunc(content)
			stop = start
		} else {
			start = 0
			stop = 0
		}
	} else if t.skipHeaderFunc != nil {
		start += t.skipHeaderFunc(content[start:stop])
	}
	if year == "" {
		year = fmt.Sprint(time.Now().Year())
	}

	return content[:start], content[start:stop], content[stop:], year
}

// Look for a boilerplate block (C/C++/Shell-style comment, contains boilerplate keywords),
// and return its start/end byte index.
func findExistingBoilerplate(content string) (start int, stop int, year string) {
	inblock := ""
	isBoiler := false
	pos := 0
	start = -1
	year = ""
	for line := range strings.Lines(content) {
		l := strings.TrimSpace(line)
		// Check if current line is from a boilerplate, and remember the year
		yearmatch := CopyrightRegex.FindStringSubmatch(l)
		if len(yearmatch) == 2 {
			isBoiler = true
			year = yearmatch[1]
		}

		switch inblock {
		// Check for the begining of a comment block
		case "":
			if strings.HasPrefix(l, "/*") {
				inblock = "/*"
				start = pos
				if strings.HasSuffix(l, "*/") {
					if isBoiler {
						return start, pos + len(line), year
					}
					inblock = ""
					start = -1
					isBoiler = false
				}
			} else if strings.HasPrefix(l, "//") {
				inblock = "//"
				start = pos
			} else if strings.HasPrefix(l, "#") {
				inblock = "#"
				start = pos
			}
		// Check for the end of a comment block (previous line)
		case "//", "#":
			if !strings.HasPrefix(l, inblock) {
				inblock = ""
				if start >= 0 && isBoiler {
					return start, pos, year
				}
				start = -1
				isBoiler = false
			}
		// Check for the end of a comment block (current line)
		case "/*":
			if strings.HasSuffix(l, "*/") {
				inblock = ""
				if start >= 0 && isBoiler {
					return start, pos + len(line), year
				}
				start = -1
				isBoiler = false
			}
		}
		pos += len(line)
	}
	// Handle "boilerplate reaches end of file" case
	if inblock != "" && start >= 0 && isBoiler {
		return start, pos, year
	}
	return -1, -1, ""
}

// Find location past the go build constraints
func skipHeaderGoFile(raw string) int {
	loc := BuildConstraintsRegex.FindStringIndex(raw)
	if loc != nil {
		return loc[1]
	}
	return 0
}

// Find location past the shebang line
func skipHeaderShebang(raw string) int {
	loc := ShebangRegex.FindStringIndex(raw)
	if loc != nil {
		return loc[1]
	}
	return 0
}
