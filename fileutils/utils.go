package fileutils

import "os"

func IsFileExist(filename string) bool {
	_, err := os.Stat(filename)
	if err == nil {
		return true
	}
	return os.IsExist(err)
}
