package sessionNo

import "strings"

// stine urls are from the following format
// dispatcher + "?APPNAME=" + applicationName + "&PRGNAME=" + programName + "&ARGUMENTS=-N" + sessionNo + ",-N" + menuId  + temp_args
func Get(urlStr string) string {
	_, path, _ := strings.Cut(urlStr, "ARGUMENTS=-N")
	return path[:15]
}
