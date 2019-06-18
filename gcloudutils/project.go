package gcloudutils

import "os"

var ProjectName = ""

func GetProjectId() string {
	if ProjectName == "" {
		proj := os.Getenv("PROJECT_ID")
		if proj == "" {
			proj = "solo-public"
		}
		ProjectName = proj
	}
	return ProjectName
}
