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

// BoilerplateTemplate takes a raw template as input and pre-processes it so it's ready for use
// during validation.
type BoilerplateTemplate struct {
	text           string
	skipHeaderFunc func(string) int
}

// BoilerplateTemplateConfiguration holds configuration values which can be used for pre-processing a template
type BoilerplateTemplateConfiguration struct {
	// ExpectedAuthor contains the name of the author expected to be found in
	// the template. Related to the <<AUTHOR>> marker.
	ExpectedAuthor string

	// SkipHeaderFunc is an optional parsing step for files matched by this template.
	// For example, in go files the boilerplate should go after build constraints.
	SkipHeaderFunc func(string) int
}

// NewBoilerplateTemplate creates a new boilerplate template using the given raw template and configuration
func NewBoilerplateTemplate(raw string, config BoilerplateTemplateConfiguration) (BoilerplateTemplate, error) {
	if !strings.Contains(raw, CopyrightMarker) {
		return BoilerplateTemplate{}, fmt.Errorf("couldn't find replacement marker %q", CopyrightMarker)
	}

	if !strings.Contains(raw, AuthorMarker) {
		return BoilerplateTemplate{}, fmt.Errorf("couldn't find replacement marker %q", AuthorMarker)
	}

	text := strings.ReplaceAll(raw, AuthorMarker, config.ExpectedAuthor)
	text = strings.TrimSpace(text) + "\n"

	return BoilerplateTemplate{
		text:           text,
		skipHeaderFunc: config.SkipHeaderFunc,
	}, nil
}

// Validate checks the given file path against the template
func (t BoilerplateTemplate) Validate(path string, patch bool) error {
	// Read file and check
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read: %w", err)
	}

	if SkipFileRegex.Match(content) || GeneratedRegex.Match(content) {
		return nil
	}

	// Find boilerplate year and location, make sure we have exactly one newline around the boilerplate
	head, boilOrig, foot, year := t.analyzeFile(content)
	have := head + boilOrig + foot
	boilExpect := strings.ReplaceAll(t.text, YearMarker, year)
	if head != "" {
		head = strings.TrimSpace(head) + "\n\n"
	}
	if foot != "" {
		foot = "\n" + strings.TrimLeftFunc(foot, unicode.IsSpace)
	}
	want := head + boilExpect + foot

	// Return error and patch if we don't have what we want
	if have != want {
		reason := "incorrect boilerplate"
		if boilOrig == "" {
			reason = "missing boilerplate"
		}
		if patch {
			edits := myers.ComputeEdits(span.URIFromPath(path), have, want)
			return fmt.Errorf("%s\n%s", reason, gotextdiff.ToUnified(path, "expected", have, edits))
		} else {
			return fmt.Errorf("%s", reason)
		}
	}

	return nil
}

// Split the input into header/boilerplate/footer parts, and finds the copyright year.
// The boilerplate part may be empty, and in this case the copyright year is generated.
// The header might be a shebang, golang build constraints, etc (see LoadTemplates).
func (t BoilerplateTemplate) analyzeFile(raw []byte) (head string, boil string, foot string, year string) {
	// Remove any windows-style line feeds in the raw input
	content := strings.ReplaceAll(string(raw), "\r", "")

	// Find location/year of existing boilerplate, or generate one
	start, stop, year := findExistingBoilerplate(content)
	if start == -1 {
		year = fmt.Sprint(time.Now().Year())
		if t.skipHeaderFunc != nil {
			start = t.skipHeaderFunc(content)
			stop = start
		} else {
			start = 0
			stop = 0
		}
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
