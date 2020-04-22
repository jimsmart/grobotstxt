// Copyright 2020 Jim Smart
// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// This file tests the robots.txt parsing and matching code found in robots.cc
// against the current Robots Exclusion Protocol (REP) internet draft (I-D).
// https://tools.ietf.org/html/draft-koster-rep

//

// Converted 2020-04-21, from https://github.com/google/robotstxt/blob/master/robots_test.cc

package grobotstxt

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Robots private", func() {

	// Line :948
	TestPath := func(uri, expected string) {
		x := getPathParamsQuery(uri)
		Expect(x).To(Equal(expected))
	}

	TestEscape := func(uri, expected string) {
		x := escapePattern(uri)
		Expect(x).To(Equal(expected))
	}

	It("should TestGetPathParamsQuery", func() {
		// Only testing URLs that are already correctly escaped here.
		TestPath("", "/")
		TestPath("http://www.example.com", "/")
		TestPath("http://www.example.com/", "/")
		TestPath("http://www.example.com/a", "/a")
		TestPath("http://www.example.com/a/", "/a/")
		TestPath("http://www.example.com/a/b?c=http://d.e/", "/a/b?c=http://d.e/")
		TestPath("http://www.example.com/a/b?c=d&e=f#fragment", "/a/b?c=d&e=f")
		TestPath("example.com", "/")
		TestPath("example.com/", "/")
		TestPath("example.com/a", "/a")
		TestPath("example.com/a/", "/a/")
		TestPath("example.com/a/b?c=d&e=f#fragment", "/a/b?c=d&e=f")
		TestPath("a", "/")
		TestPath("a/", "/")
		TestPath("/a", "/a")
		TestPath("a/b", "/b")
		TestPath("example.com?a", "/?a")
		TestPath("example.com/a;b#c", "/a;b")
		TestPath("//a/b/c", "/b/c")
	})

	It("should TestMaybeEscapePattern", func() {
		TestEscape("http://www.example.com", "http://www.example.com")
		TestEscape("/a/b/c", "/a/b/c")
		TestEscape("á", "%C3%A1")
		TestEscape("%aa", "%AA")
	})

})
