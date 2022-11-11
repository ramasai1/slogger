// Copyright 2013 MongoDB, Inc.
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

// Package rolling_file_appender provides a slogger Appender that
// supports log rotation.

package test_util

import (
	"strings"
	"testing"
)

func AssertNoErrors(test *testing.T, errs []error) {
	if len(errs) != 0 {
		test.Errorf("Expected to be empty: %v", errs)
	}
}

func AssertErrorExists(test *testing.T, errs []error, errorStr string) {
	foundErr := false
	for _, err := range errs {
		if strings.Compare(err.Error(), errorStr) == 0 {
			foundErr = true
			break
		}
	}

	if !foundErr {
		test.Errorf("Expected errorStr to be part of the errors slice.: %v, %v", errs, errorStr)
	}
}
