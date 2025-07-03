package utils

import (
	"fmt"
	"os"
	"strconv"
)

func getFileVersion(path string) string {
	info, err := os.Stat(fmt.Sprintf("../%s", path))
	if err != nil {
		return "0"
	}
	return strconv.FormatInt(info.ModTime().Unix(), 10)
}

func AssetURL(path string) string {
	return fmt.Sprintf("%s?v=%s", path, getFileVersion(path))
}
