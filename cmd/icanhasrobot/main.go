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
// -----------------------------------------------------------------------------
// File: robots_main.cc
// -----------------------------------------------------------------------------
//
// Simple binary to assess whether a URL is accessible to a set of user-agents
// according to records found in a local robots.txt file, based on Google's
// robots.txt parsing and matching algorithms.
// Usage:
//     robots_main <local_path_to_robotstxt> <user_agents> <url>
// Arguments:
// local_path_to_robotstxt: local path to a file containing robots.txt records.
//   For example: /home/users/username/robots.txt
// user_agents: a token to be matched against records in the robots.txt (or
// a comma-separated list of user agents)
//   For example: Googlebot or Googlebot,Googlebot-image
// url: a url to be matched against records in the robots.txt. The URL must be
// %-encoded according to RFC3986.
//   For example: https://example.com/accessible/url.html
// Returns: Prints a sentence with verdict about whether 'user_agent' is allowed
// to access 'url' based on records in 'local_path_to_robotstxt'. Exits with status
// code 0 if allowed, 1 if disallowed, or 2 otherwise (e.g. bad inputs).
//

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/jimsmart/grobotstxt"
)

func loadFile(filename string) (string, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// TODO(js) Can we use the flag package here, instead of manually handling params and help/usage.

func showHelp(argv []string) {
	fmt.Fprint(os.Stderr, "Shows whether the given user_agent(s) and URI combination"+
		" is allowed or disallowed by the given robots.txt file.\n\n")
	fmt.Fprint(os.Stderr, "Usage:\n"+
		"  "+argv[0]+" <robots.txt filename> <user_agents> <URI>\n\n")
	// fmt.Fprint(os.Stderr, "The URI must be %-encoded according to RFC3986.\n\n")
	fmt.Fprint(os.Stderr, "The user_agents may be a single UA or a comma-separated string.\n\n")
	fmt.Fprint(os.Stderr, "Example:\n"+
		"  "+argv[0]+" robots.txt FooBot http://example.com/foo\n")
}

func main() {
	argv := os.Args

	filename := ""
	if len(argv) >= 2 {
		filename = argv[1]
	}
	if filename == "-h" || filename == "-help" || filename == "--help" {
		showHelp(argv)
		os.Exit(2)
	}

	if len(argv) != 4 {
		fmt.Fprint(os.Stderr, "Invalid amount of arguments. Showing help.\n\n")
		showHelp(argv)
		os.Exit(2)
	}

	robotsContent, err := loadFile(filename)
	if err != nil {
		fmt.Fprint(os.Stderr, "failed to read file \""+filename+"\"\n")
		os.Exit(2)
	}

	userAgent := argv[2]
	userAgentList := strings.Split(userAgent, ",")
	uri := argv[3]

	allowed := grobotstxt.AgentsAllowed(robotsContent, userAgentList, uri)

	m := "user-agent '" + userAgent + "' with URI '" + uri + "': "
	if allowed {
		m += "ALLOWED"
	} else {
		m += "DISALLOWED"
	}
	m += "\n"
	fmt.Fprint(os.Stdout, m)
	if len(robotsContent) == 0 {
		fmt.Fprint(os.Stdout, "notice: robots file is empty so all user-agents are allowed\n")
	}

	if allowed {
		// Return 0 if URL is allowed for user-agent.
		os.Exit(0)
	} else {
		// Return 1 if URL is disallowed for user-agent.
		os.Exit(1)
	}
}
