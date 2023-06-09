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
	"testing"
)

func Test_ShebangRegex(t *testing.T) {
	tests := map[string]struct {
		input       string
		shouldMatch bool
	}{
		"sh shebang": {
			shouldMatch: true,
			input:       "#!/bin/sh\n",
		},
		"python3 shebang": {
			shouldMatch: true,
			input:       "#!/usr/bin/env python3\n",
		},
		"python2 shebang": {
			shouldMatch: true,
			input:       "#!/usr/bin/env python2\n",
		},
		"python shebang": {
			shouldMatch: true,
			input:       "#!/usr/bin/env python\n",
		},
		"bash shebang": {
			shouldMatch: true,
			input:       "#!/usr/bin/env bash\n",
		},
		"no shebang": {
			shouldMatch: false,
			input:       "package main\n\nfunc main() {}",
		},
		"no newline on shebang": {
			shouldMatch: false,
			input:       "#!/bin/sh",
		},
		"many newlines on shebang": {
			shouldMatch: true,
			input:       "#!/bin/sh\n\n\n\n\ntest",
		},
		"longer file python shebang": {
			shouldMatch: true,
			input:       skipFileLong,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			matched := ShebangRegex.MatchString(test.input)

			if matched != test.shouldMatch {
				t.Errorf("matched=%v, shouldMatch=%v", matched, test.shouldMatch)
			}
		})
	}
}

func Test_BuildConstraintsRegex(t *testing.T) {
	tests := map[string]struct {
		input       string
		shouldMatch bool
	}{
		"old style build constraint": {
			shouldMatch: true,
			input:       "// +build linux\n",
		},
		"new style build constraint": {
			shouldMatch: true,
			input:       "//go:build linux\n",
		},
		"both styles of build constraint": {
			shouldMatch: true,
			input:       "// +build linux\n//go:build linux\n",
		},
		"longer non-go file without build constraints": {
			shouldMatch: false,
			input:       skipFileLong,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			matched := BuildConstraintsRegex.MatchString(test.input)

			if matched != test.shouldMatch {
				t.Errorf("matched=%v, shouldMatch=%v", matched, test.shouldMatch)
			}
		})
	}
}

func Test_GeneratedRegex(t *testing.T) {
	tests := map[string]struct {
		input       string
		shouldMatch bool
	}{
		"MockGen": {
			// from kubernetes/kubernetes
			shouldMatch: true,
			input:       "// Code generated by MockGen. DO NOT EDIT.\n",
		},
		"swagger": {
			// from kubernetes/kubernetes
			shouldMatch: true,
			input:       "// AUTO-GENERATED FUNCTIONS START HERE. DO NOT EDIT.\n",
		},
		"defaulter-gen": {
			shouldMatch: true,
			input:       "// Code generated by defaulter-gen. DO NOT EDIT.\n",
		},
		"deepcopy-gen": {
			shouldMatch: true,
			input:       "// Code generated by deepcopy-gen. DO NOT EDIT.\n",
		},
		"conversion-gen": {
			shouldMatch: true,
			input:       "// Code generated by conversion-gen. DO NOT EDIT.\n",
		},
		"client-gen": {
			shouldMatch: true,
			input:       "// Code generated by client-gen. DO NOT EDIT.\n",
		},
		"informer-gen": {
			shouldMatch: true,
			input:       "// Code generated by informer-gen. DO NOT EDIT.\n",
		},
		"informer-gen but in python": {
			shouldMatch: true,
			input:       "# Code generated by informer-gen. DO NOT EDIT.\n",
		},
		"longer file with no matches": {
			shouldMatch: false,
			input:       skipFileLong,
		},
		"longer file with a match": {
			shouldMatch: true,
			input:       generatedFileLong,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			matched := GeneratedRegex.MatchString(test.input)

			if matched != test.shouldMatch {
				t.Errorf("matched=%v, shouldMatch=%v", matched, test.shouldMatch)
			}
		})
	}
}

func Test_SkipFileRegex(t *testing.T) {
	tests := map[string]struct {
		input       string
		shouldMatch bool
	}{
		"golang style comment": {
			shouldMatch: true,
			input:       "// +skip_license_check\n",
		},
		"python / bash style comment": {
			shouldMatch: true,
			input:       "# +skip_license_check\n",
		},
		"longer file": {
			shouldMatch: true,
			input:       skipFileLong,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			matched := SkipFileRegex.MatchString(test.input)

			if matched != test.shouldMatch {
				t.Errorf("matched=%v, shouldMatch=%v", matched, test.shouldMatch)
			}
		})
	}
}

const (
	// skipFileLong is for testing SkipFileRegex; should match also for the
	// shebang regex, but most of the others shouldn't match
	skipFileLong = `#!/usr/bin/env python

# +skip_license_check

# Copyright 2015 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Verifies that all source files contain the necessary copyright boilerplate
# snippet.

from __future__ import print_function

import argparse
import datetime
`

	// generatedFileLong should match for GeneratedRegex
	generatedFileLong = `/*
Copyright The cert-manager Authors.

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

// Code generated by lister-gen. DO NOT EDIT.
`
)
