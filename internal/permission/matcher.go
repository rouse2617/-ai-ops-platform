package permission

import "strings"

func Match(pattern, name string) bool {
	return matchPattern(pattern, name)
}

func matchPattern(pattern, name string) bool {
	if pattern == "**" {
		return true
	}

	pi, ni := 0, 0
	plen, nlen := len(pattern), len(name)
	star := -1
	match := 0

	for ni < nlen {
		if pi < plen && (pattern[pi] == name[ni] || pattern[pi] == '?') {
			pi++
			ni++
		} else if pi < plen && pattern[pi] == '*' {
			star = pi
			match = ni
			pi++
		} else if star != -1 {
			pi = star + 1
			match++
			ni = match
		} else {
			return false
		}
	}

	for pi < plen && pattern[pi] == '*' {
		pi++
	}

	return pi == plen
}

func matchGlob(pattern, name string) bool {
	if pattern == name {
		return true
	}

	if !strings.Contains(pattern, "*") && !strings.Contains(pattern, "?") {
		return pattern == name
	}

	return matchPattern(pattern, name)
}
