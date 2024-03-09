package sessionNo

import "testing"

func TestRefreshSessionNumber(t *testing.T) {
	fakeRegistrationLink := "https://stine.uni-hamburg.de/scripts/mgrqispi.dll?APPNAME=CampusNet&PRGNAME=REGCOURSEMOD&ARGUMENTS=-N232343443351119,-N343449,-N343424234011169,-ADOFF,-N343434342285453,-N344343434341730,-N0,-N0,-N0,-AN,-N0"
	newSessionNo := "232343443351119"

	urlWithRefreshedSessionNo := Refresh(fakeRegistrationLink, newSessionNo)

	if urlWithRefreshedSessionNo != "https://stine.uni-hamburg.de/scripts/mgrqispi.dll?APPNAME=CampusNet&PRGNAME=REGCOURSEMOD&ARGUMENTS=-N"+newSessionNo+",-N343449,-N343424234011169,-ADOFF,-N343434342285453,-N344343434341730,-N0,-N0,-N0,-AN,-N0" {
		t.Error("session number is not being replaced in link")
	}
}
