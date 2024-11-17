package helmutils

import (
	"strings"
)

/*
	FindHelmChartWhiteSpaces is used to find unusual spaces in helm charts, that could cause issues.
	For example, if a line is indented by 2 spaces, and the next line is indented by 6 spaces, this could cause issues.
	The line would need to be indented by 4 spaces instead.

	Please check the tests for what is considered a good and bad formatted yaml section.

	List of features
	- looks for spacing issues with the next line is off by more than 2
	- if next line is an array, it can be off by 4 or 2 spaces.
	- if the line has special breaks (|-, >-, etc.) identified in YAML,
		these are now ignored areas in the YAML.
	- if the YAML has empty lines, these are acceptable.
	- if the YAMl has spaces in an empty line, there is an option to control
		whether to see these or not. Use the DetectWhiteSpacesInEmptyLines option when parsing.
*/

// returns the windows of the helm chart that contain white spacing and formatting issues.
func FindHelmChartWhiteSpaces(data string, opts HelmDetectOptions) [][]string {
	lines := strings.Split(string(data), "\n")
	// we want to count each line, if the number of spaces at the beginning is equal to 0, +2, or -2 from the previous line
	// then we want to continue to the next line. Else we want to throw an error.
	previous := previousInfo{NumOfSpaces: 0, BeganWithArray: false}
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
		// next level is the next acceptable number of spaces
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
		// // this means an empty line has occurred, and it contains only spaces, so move on to the next line
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
			previous = previousInfo{NumOfSpaces: currentNumOfSpaces + 2, BeganWithArray: beginsWithArray}
		} else {
			previous = previousInfo{NumOfSpaces: currentNumOfSpaces, BeganWithArray: beginsWithArray}
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

// NewSpaces will return a new Spaces struct
func NewSpaces(s string) *Spaces {
	spaces := strings.Split(s, " ")
	return &Spaces{line: s, spaces: spaces}
}

// Spaces is a struct that contains information about a line
// and the spaces after the line has been split on a space.
type Spaces struct {
	line   string
	spaces []string
}

// GetNumberOfSpacesAtBeginning returns the number of spaces (' ') at the beginning of the line
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

// BeginsWithArray returns if the line starts as an YAML array
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

// IsNewResource returns if the line starts a new resource
func (s *Spaces) IsNewResource() bool {
	if len(s.spaces) == 0 {
		return false
	}
	return s.spaces[0] == "---"
}

// StartsWithComment returns if the line starts with a comment
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

// IsEmptyLine returns true if the line is empty
func (s *Spaces) IsEmptyLine() bool {
	return s.line == ""
}

// HasSpecialBreak if the last section contains a special break
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

// ContainsOnlySpaces returns true if the line only contains spaces
func (s *Spaces) ContainsOnlySpaces() bool {
	for _, space := range s.spaces {
		if space != "" {
			return false
		}
	}
	return true
}

// previousInfo is used to keep track of the previous line
type previousInfo struct {
	NumOfSpaces    int
	BeganWithArray bool
}

// HelmDetectOptions is used to configure the FindHelmChartWhiteSpaces function
type HelmDetectOptions struct {
	// DetectWhiteSpacesInEmptyLines is used to detect white spaces in empty lines
	DetectWhiteSpacesInEmptyLines bool
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
