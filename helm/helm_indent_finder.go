package helm

import (
	"strings"
)

func NewSpaces(s string) *Spaces {
	spaces := strings.Split(s, " ")
	return &Spaces{line: s, spaces: spaces}
}

type Spaces struct {
	line   string
	spaces []string
}

// Returns the number of spaces (' ') at the beginning of the line
func (s *Spaces) GetNumberOfSpacesAtBeginning() int {
	numberOfSpaces := 0
	for _, space := range s.spaces {
		if space == "" {
			numberOfSpaces++
		} else {
			break
		}
	}
	if s.ContainsOnlySpaces() {
		return numberOfSpaces - 1
	}
	return numberOfSpaces
}

// returns if the line starts as an YAML array
func (s *Spaces) BeginsWithArray() bool {
	for _, space := range s.spaces {
		if space == "" {
			continue
		} else if space == "-" {
			return true
		} else {
			return false
		}
	}
	return false
}

// returns if the line starts a new resource
func (s *Spaces) IsNewResource() bool {
	if len(s.spaces) == 0 {
		return false
	}
	return s.spaces[0] == "---"
}

// returns if the line starts with a comment
func (s *Spaces) StartsWithComment() bool {
	if len(s.spaces) == 0 {
		return false
	}
	// continue until the first space contains a comment, or contains anything that is not a comment
	for _, space := range s.spaces {
		if strings.HasPrefix(space, "#") {
			return true
		} else if space == "" {
			continue
		} else {
			return false
		}
	}
	return false
}

func (s *Spaces) IsEmptyLine() bool {
	return s.line == ""
}

// if the last section contians a bar dash (|-)
func (s *Spaces) HasSpecialBreak() bool {
	specialBreaks := []string{"|", "|-", "|+", ">", ">+", ">-"}
	chars := s.spaces[len(s.spaces)-1]
	for _, specialBreak := range specialBreaks {
		if chars == specialBreak {
			return true
		}
	}
	return false
}

// returns true if the line only contains spaces
func (s *Spaces) ContainsOnlySpaces() bool {
	for _, space := range s.spaces {
		if space != "" {
			return false
		}
	}
	return true
}

type PreviousInfo struct {
	NumOfSpaces    int
	BeganWithArray bool
}

type HelmDetectOptions struct {
	// DetectWhiteSpacesInEmptyLines is used to detect white spaces in empty lines
	DetectWhiteSpacesInEmptyLines bool
}

// returns the windows of the helm chart that contain white spacing and formatting issues.
func FindHelmChartWhiteSpaces(data string, opts HelmDetectOptions) [][]string {
	lines := strings.Split(string(data), "\n")
	// we want to count each line, if the number of spaces at the begining is equal to 0, +2, or -2 from the previous line
	// then we want to continue to the next line. Else we want to throw an error.
	previous := PreviousInfo{NumOfSpaces: 0, BeganWithArray: false}
	badWindows := [][]string{}
	specialBreak := false
	for currentIndex, line := range lines {
		s := NewSpaces(line)
		startsWithComment := s.StartsWithComment()
		if startsWithComment {
			continue
		}
		isEmptyLine := s.IsEmptyLine()
		// if the line is empty we do not care about it
		if isEmptyLine {
			continue
		}
		containsOnlySpaces := s.ContainsOnlySpaces()
		// if we do not detect white spaces continue
		if !opts.DetectWhiteSpacesInEmptyLines && containsOnlySpaces {
			continue
		}
		shouldContinue := false
		currentNumOfSpaces := s.GetNumberOfSpacesAtBeginning()
		beginsWithArray := s.BeginsWithArray()
		// next level is the next accpetable number of spaces
		nextLevel := previous.NumOfSpaces + 2
		twoLevels := previous.NumOfSpaces + 4
		isCurrentLevel := previous.NumOfSpaces == currentNumOfSpaces
		isTwoLevelsAway := currentNumOfSpaces == twoLevels
		isNextLevel := currentNumOfSpaces == nextLevel
		isSmallerLevel := currentNumOfSpaces < previous.NumOfSpaces

		if isSmallerLevel && specialBreak {
			// we are now exiting the barDash
			specialBreak = false
		} else if specialBreak {
			// if we are in a specialBreak, just continue regardless,
			// until we exit the specialBreak
			continue
		}
		// // this means an empty line has occured, and it contains only spaces, so move on to the next line
		if (isCurrentLevel || isNextLevel) && containsOnlySpaces {
			continue
		}

		if isCurrentLevel || isNextLevel || isSmallerLevel {
			shouldContinue = true
			// if the current is less than previous, we are moving out of an object or array
			// if it is an empty line we need to ignore it, nothing happens
		} else if previous.BeganWithArray && isTwoLevelsAway {
			shouldContinue = true
		}

		// always set the previous number of spaces, regardless
		if beginsWithArray && (isCurrentLevel || isSmallerLevel) {
			// add 2 to current level because the array can be on current level or next level
			// if on current level, the next line should be on the next level
			previous = PreviousInfo{NumOfSpaces: currentNumOfSpaces + 2, BeganWithArray: beginsWithArray}
		} else {
			previous = PreviousInfo{NumOfSpaces: currentNumOfSpaces, BeganWithArray: beginsWithArray}
		}

		// just record that there is a bar dash
		if s.HasSpecialBreak() {
			specialBreak = true
		}
		if shouldContinue {
			continue
		} else {
			windowLines := Window(lines, currentIndex, 6)
			badWindows = append(badWindows, windowLines)
		}
	}
	return badWindows
}

func Window(s []string, index int, windowSize int) []string {
	start := index - windowSize
	end := index + windowSize
	if len(s) < end {
		end = len(s)
	}
	if start < 0 {
		start = 0
	}
	return s[start:end]
}
