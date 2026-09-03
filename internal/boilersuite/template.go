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
)

// BoilerplateTemplate takes a raw template as input and pre-processes it so it's ready for use
// during validation.
type BoilerplateTemplate struct {
	raw      string
	replaced string

	lineCount int

	normalizationFunc func(string) string
}

// BoilerplateTemplateConfiguration holds configuration values which can be used for pre-processing a template
type BoilerplateTemplateConfiguration struct {
	// ExpectedAuthor contains the name of the author expected to be found in
	// the template. Related to the <<AUTHOR>> marker.
	ExpectedAuthor string

	// NormalizationFunc is an optional extra normalization step to take for
	// files matched by this template. For example, in go files we might need
	// to remove golang build constraints
	NormalizationFunc func(string) string
}

// NewBoilerplateTemplate creates a new boilerplate template using the given raw template and configuration
func NewBoilerplateTemplate(raw string, config BoilerplateTemplateConfiguration) (BoilerplateTemplate, error) {
	if !strings.Contains(raw, YearMarker) {
		return BoilerplateTemplate{}, fmt.Errorf("couldn't find replacement marker %q", YearMarker)
	}

	if !strings.Contains(raw, AuthorMarker) {
		return BoilerplateTemplate{}, fmt.Errorf("couldn't find replacement marker %q", AuthorMarker)
	}

	replaced := strings.ReplaceAll(raw, AuthorMarker, config.ExpectedAuthor)

	lineCount := strings.Count(replaced, "\n") + 1

	return BoilerplateTemplate{
		raw:               raw,
		replaced:          replaced,
		lineCount:         lineCount,
		normalizationFunc: config.NormalizationFunc,
	}, nil
}

// Validate checks the given file path against the template
func (t BoilerplateTemplate) Validate(path string) error {
	contents, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read: %w", err)
	}

	if SkipFileRegex.Match(contents) || GeneratedRegex.Match(contents) {
		return nil
	}

	normalizedContents, err := t.normalizeAndTrimFile(string(contents))
	if err != nil {
		return err
	}

	if !strings.HasPrefix(normalizedContents, t.replaced) {
		return fmt.Errorf("does not start with expected template type")
	}

	return nil
}

// normalizeAndTrimFile takes a given input file and strips any shebang lines,
// Golang build constraints and any leading or trailing whitespace
func (t BoilerplateTemplate) normalizeAndTrimFile(raw string) (string, error) {
	raw = strings.ReplaceAll(raw, "\r", "")

	raw = fileBeginning(raw, t.lineCount)

	if t.normalizationFunc != nil {
		raw = t.normalizationFunc(raw)
	}

	// replace anything which looks like a date with the year marker
	raw = DateRegex.ReplaceAllString(raw, CopyrightMarker)

	// Remove any windows-style line feeds in the raw input

	raw = strings.TrimLeft(raw, "\n")

	split := strings.Split(raw, "\n")

	if len(split) < t.lineCount {
		return raw, fmt.Errorf("file is shorter than the boilerplate header; cannot have correct boilerplate")
	}

	return strings.Join(split[:t.lineCount], "\n"), nil
}

func fileBeginning(raw string, templateLineCount int) string {
	s := strings.Split(raw, "\n")
	if len(s) >= templateLineCount*2 {
		s = s[:templateLineCount*2]
	}

	return strings.Join(s, "\n")
}

func normalizeGoFile(raw string) string {
	// Remove any golang build constraints
	return BuildConstraintsRegex.ReplaceAllString(raw, "")
}

func normalizeShebang(raw string) string {
	// Remove the shebang line, if there is one
	return ShebangRegex.ReplaceAllString(raw, "")
}
