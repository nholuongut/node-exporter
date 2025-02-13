// Copyright 2015 The Nho Luong Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build !nofilefd
// +build !nofilefd

package collector

import "testing"

func TestFileFDStats(t *testing.T) {
	fileFDStats, err := parseFileFDStats("fixtures/proc/sys/fs/file-nr")
	if err != nil {
		t.Fatal(err)
	}

	if want, got := "1024", fileFDStats["allocated"]; want != got {
		t.Errorf("want filefd allocated %q, got %q", want, got)
	}

	if want, got := "1631329", fileFDStats["maximum"]; want != got {
		t.Errorf("want filefd maximum %q, got %q", want, got)
	}
}
