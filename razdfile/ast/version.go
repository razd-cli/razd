package ast

// SupportedVersion is the current Razdfile schema version
const SupportedVersion = "1"

// SupportedVersions lists all supported Razdfile versions
var SupportedVersions = []string{"1"}

// IsVersionSupported checks if the given version is supported
func IsVersionSupported(version string) bool {
	for _, v := range SupportedVersions {
		if v == version {
			return true
		}
	}
	return false
}
