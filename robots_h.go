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

// ParseHandler is a handler for directives found in robots.txt.
// These callbacks are called by Parse() in the sequence they
// have been found in the file.
type ParseHandler interface {
	HandleRobotsStart()
	HandleRobotsEnd()
	HandleUserAgent(lineNum int, value string)
	HandleAllow(lineNum int, value string)
	HandleDisallow(lineNum int, value string)
	HandleSitemap(lineNum int, value string)
	HandleUnknownAction(lineNum int, action, value string)
}

var _ ParseHandler = &RobotsMatcher{}

// RobotsMatcher — matches robots.txt against URIs.
//
// The RobotsMatcher uses a default match strategy for Allow/Disallow patterns which
// is the official way of Google crawler to match robots.txt. It is also
// possible to provide a custom match strategy.
//
// The entry point for the user is to call one of the AgentAllowed()
// methods that return directly if a URI is being allowed according to the
// robots.txt and the crawl agent.
//
// The RobotsMatcher can be re-used for URIs/robots.txt but is not concurrency-safe.
type RobotsMatcher struct {
	// Line :87

	// Line :222

	allow    *matchHierarchy // Characters of 'uri' matching Allow.
	disallow *matchHierarchy // Characters of 'uri' matching Disallow.

	seenGlobalAgent       bool // True if processing global agent rules.
	seenSpecificAgent     bool // True if processing our specific agent.
	everSeenSpecificAgent bool // True if we ever saw a block for our agent.
	seenSeparator         bool // True if saw any key: value pair.

	// The path we want to pattern match.
	path string

	// The User-Agents we are interested in.
	userAgents []string

	MatchStrategy MatchStrategy
}

func (m *RobotsMatcher) seenAnyAgent() bool {
	// Line :167
	return m.seenGlobalAgent || m.seenSpecificAgent
}

//

const noMatchPriority = -1

// Instead of just maintaining a Boolean indicating whether a given line has
// matched, we maintain a count of the maximum number of characters matched by
// that pattern.
//
// This structure stores the information associated with a match (e.g. when a
// Disallow is matched) as priority of the match and line matching.
//
// The priority is initialized with a negative value to make sure that a match
// of priority 0 is higher priority than no match at all.
type match struct {
	// Line :181
	priority int
	line     int
}

// newMatch returns a new Match with an initial priority of noMatchPriority.
func newMatch() *match {
	return &match{
		priority: noMatchPriority,
	}
}

func (m *match) Set(priority, line int) {
	m.priority = priority
	m.line = line
}

// Clear resets the internal Match state
// values to their initial state.
func (m *match) Clear() {
	m.Set(noMatchPriority, 0)
}

// higherPriorityMatch takes two Matches and returns
// the one with the highest priority.
func higherPriorityMatch(a, b *match) *match {
	if a.priority > b.priority {
		return a
	}
	return b
}

//

// For each of the directives within user-agent sections, we keep global and specific
// match scores.

// Line :212

type matchHierarchy struct {
	global   *match // Match for '*'
	specific *match // Match for queried agent.
}

func newMatchHierarchy() *matchHierarchy {
	return &matchHierarchy{
		global:   newMatch(),
		specific: newMatch(),
	}
}

func (m *matchHierarchy) Clear() {
	m.global.Clear()
	m.specific.Clear()
}
