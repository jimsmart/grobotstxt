package grobotstxt

func IsValidUserAgentToObey(userAgent string) bool {
	return NewRobotsMatcher().isValidUserAgentToObey(userAgent)
}
