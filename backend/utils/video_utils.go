package utils

import (
	"os"
)

// BuildFinalConcatList returns a list of video paths to be concatenated.
// It handles intro/outro logic based on platform and file existence.
func BuildFinalConcatList(platform, introPath, outroPath, mainVideoPath string) []string {
	var concatList []string

	if platform == "youtube" {
		if _, err := os.Stat(introPath); err == nil {
			concatList = append(concatList, introPath)
		}
	}

	concatList = append(concatList, mainVideoPath)

	if platform == "youtube" {
		if _, err := os.Stat(outroPath); err == nil {
			concatList = append(concatList, outroPath)
		}
	}

	return concatList
}
