// Copyright 2024 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

/*
This file contains tests for the getSupportedTypes function and provides a
helper to print out supported types to add to documentation
*/
package configurature

import (
	"flag"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// For getSupportedTypes test helper to print out supported types to add to
// documentation
var printTypes = flag.Bool("print_types", false, "print supported types")

func TestGetSupportedTypes(t *testing.T) {
	flag.Parse()
	if *printTypes {
		for _, v := range getSupportedTypes() {
			fmt.Printf("type=%s\n", v)
		}
	} else {
		assert.GreaterOrEqual(t, len(getSupportedTypes()), 30)
	}
}
