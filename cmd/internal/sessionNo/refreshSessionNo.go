package sessionNo

import "regexp"

/*
Refresh replaces the session number in a STiNE url with the session number of the current [Session].
The session number in the URL needs to correspond with a specific cookie to authenticate on STiNE.
*/
func Refresh(url string, sessionNo string) string {
	reg := regexp.MustCompile("ARGUMENTS=-N\\d{15}")
	return reg.ReplaceAllString(url, "ARGUMENTS=-N"+sessionNo)
}
