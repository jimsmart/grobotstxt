// Copyright 2020 Jim Smart
// Copyright 1999 Google LLC
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
// -----------------------------------------------------------------------------
// File: robots.h
// -----------------------------------------------------------------------------
//
// This file implements the standard defined by the Robots Exclusion Protocol
// (REP) internet draft (I-D).
//   https://tools.ietf.org/html/draft-koster-rep
//
// Google doesn't follow the standard strictly, because there are a lot of
// non-conforming robots.txt files out there, and we err on the side of
// disallowing when this seems intended.
//
// An more user-friendly description of how Google handles robots.txt can be
// found at:
//   https://developers.google.com/search/reference/robots_txt
//
// This library provides a low-level parser for robots.txt (ParseRobotsTxt()),
// and a matcher for URLs against a robots.txt (class RobotsMatcher).

//

// Converted 2020-04-21, from https://github.com/google/robotstxt/blob/master/robots.h

package grobotstxt

// Handler for directives found in robots.txt. These callbacks are called by
// ParseRobotsTxt() in the sequence they have been found in the file.
type RobotsParseHandler interface {
	HandleRobotsStart()
	HandleRobotsEnd()
	HandleUserAgent(lineNum int, value string)
	HandleAllow(lineNum int, value string)
	HandleDisallow(lineNum int, value string)
	HandleSitemap(lineNum int, value string)
	HandleUnknownAction(lineNum int, action, value string)
}

var _ RobotsParseHandler = &RobotsMatcher{}

// RobotsMatcher - matches robots.txt against URLs.
//
// The Matcher uses a default match strategy for Allow/Disallow patterns which
// is the official way of Google crawler to match robots.txt. It is also
// possible to provide a custom match strategy.
//
// The entry point for the user is to call one of the *AllowedByRobots()
// methods that return directly if a URL is being allowed according to the
// robots.txt and the crawl agent.
// The RobotsMatcher can be re-used for URLs/robots.txt but is not thread-safe.
type RobotsMatcher struct {
	// Line :87

	// Line :222
	allowMatch    *MatchHierarchy // Characters of 'url' matching Allow.
	disallowMatch *MatchHierarchy // Characters of 'url' matching Disallow.

	seenGlobalAgent       bool // True if processing global agent rules.
	seenSpecificAgent     bool // True if processing our specific agent.
	everSeenSpecificAgent bool // True if we ever saw a block for our agent.
	seenSeparator         bool // True if saw any key: value pair.

	// The path we want to pattern match. Not owned and only a valid pointer
	// during the lifetime of *AllowedByRobots calls.
	path string

	// The User-Agents we are interested in. Not owned and only a valid
	// pointer during the lifetime of *AllowedByRobots calls.
	userAgents []string

	matchStrategy RobotsMatchStrategy
}

func (m *RobotsMatcher) seenAnyAgent() bool {
	// Line :167
	return m.seenGlobalAgent || m.seenSpecificAgent
}

//

const NoMatchPriority = -1

// Instead of just maintaining a Boolean indicating whether a given line has
// matched, we maintain a count of the maximum number of characters matched by
// that pattern.
//
// This structure stores the information associated with a match (e.g. when a
// Disallow is matched) as priority of the match and line matching.
//
// The priority is initialized with a negative value to make sure that a match
// of priority 0 is higher priority than no match at all.
type Match struct {
	// Line :181
	priority int
	line     int
}

func (m *Match) Set(priority, line int) {
	m.priority = priority
	m.line = line
}

func (m *Match) Clear() {
	m.Set(NoMatchPriority, 0)
}

func (m *Match) HigherPriorityMatch(a, b *Match) *Match {
	return Match_HigherPriorityMatch(a, b)
}

func Match_HigherPriorityMatch(a, b *Match) *Match {
	if a.priority > b.priority {
		return a
	}
	return b
}

//

type MatchHierarchy struct {
	// Line :212
	global   *Match // Match for '*'
	specific *Match // Match for queried agent.
}

func NewMatchHierarchy() *MatchHierarchy {
	return &MatchHierarchy{
		global:   &Match{},
		specific: &Match{},
	}
}

func (m *MatchHierarchy) Clear() {
	m.global.Clear()
	m.specific.Clear()
}
