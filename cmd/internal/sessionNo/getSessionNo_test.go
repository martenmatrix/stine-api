package sessionNo

import "testing"

func TestGetSessionNo(t *testing.T) {
	sessionURL := "https://stine.uni-hamburg.de/scripts/mgrqispi.dll?APPNAME=CampusNet&PRGNAME=MLSSTART&ARGUMENTS=-N899462345432351,-N000266,"
	sessionNo := Get(sessionURL)

	if sessionNo != "899462345432351" {
		t.Errorf("WANT: 899462345432351, GOT: %s", sessionNo)
	}
}
