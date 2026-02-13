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

package main

import (
	"fmt"
	"maps"
	"slices"
	"testing"

	"github.com/cert-manager/boilersuite/internal/boilersuite"
)

func TestLoadEmbededTemplates(t *testing.T) {
	templates, err := boilersuite.LoadTemplates(boilerplateTemplateDir, "test")
	if err != nil {
		t.Fatalf("failed to load embeded templates: %s", err)
	}
	expect := "[Containerfile Dockerfile Makefile bash go mk py sh]"
	loaded := fmt.Sprintf("%s", slices.Sorted(maps.Keys(templates)))
	if loaded != expect {
		t.Fatalf("unexpected template list:\nwant %s\ngot  %s", expect, loaded)
	}
}
