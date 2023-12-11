package commands

import (
	"io/ioutil"
	"strings"

	"github.com/rotisserie/eris"
)

var (
	MalformedVersionImageConstraintLine = func(line string) error {
		return eris.Errorf("Could not properly split version image constraint line: %s", line)
	}
)

// GetImagesPerVersionFromFile Reads in a file, and tries to turn it into a map from version constraints to lists of images
// As a byproduct, it also caches all unique images found into the option field 'allImages'
// Also, I'm not sure why I didn't just use a csv reader... oh well.
func GetImagesPerVersionFromFile(constraintsFile string) (map[string][]string, error) {
	imagesPerVersion := make(map[string][]string)
	imageSet := make(map[string]interface{})

	dat, err := ioutil.ReadFile(constraintsFile)
	if err != nil {
		return nil, err
	}
	for _, line := range strings.Split(string(dat), "\n") {
		trimmedLine := strings.TrimSpace(line)
		if len(trimmedLine) == 0 || string(trimmedLine[0]) == "#" {
			continue
		}
		values := strings.Split(trimmedLine, ",")
		if len(values) < 2 {
			return nil, MalformedVersionImageConstraintLine(line)
		}
		for i, _ := range values {
			trimVal := strings.TrimSpace(values[i])
			values[i] = trimVal
			if i > 0 {
				imageSet[trimVal] = nil
			}
		}
		imagesPerVersion[values[0]] = values[1:]
	}

	return imagesPerVersion, nil
}

func GetAdditionalContextFromFile(filename string) (string, error) {
	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(contents), nil
}
