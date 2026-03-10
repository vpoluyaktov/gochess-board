package utils

import "runtime"

// GetExecutableName returns the platform-specific executable name
// On Windows, it appends .exe to the name
func GetExecutableName(name string) string {
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}
