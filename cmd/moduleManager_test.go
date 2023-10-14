package stineapi

import (
	"testing"
)

func TestCreateModuleRegistration(t *testing.T) {
	fakeRegistrationLink := "https://stine.uni-hamburg.de/scripts/mgrqispi.dll?APPNAME=CampusNet&PRGNAME=REGCOURSEMOD&ARGUMENTS=-N232343443351119,-N343449,-N343424234011169,-ADOFF,-N343434342285453,-N344343434341730,-N0,-N0,-N0,-AN,-N0"
	ses := NewSession()
	moduleReg := ses.CreateModuleRegistration(fakeRegistrationLink)
	if moduleReg.registrationLink != fakeRegistrationLink {
		t.Error("registration link is not set on object")
	}
	if moduleReg.TanRequired != false || moduleReg.TanStartsWith != "" {
		t.Error("TanRequired and TanStartsWith do not have default values")
	}
}

func TestReferenceCopyOfSessionIsUsed(t *testing.T) {
	ses := NewSession()
	moduleReg := ses.CreateModuleRegistration("https://www.example.org")

	ses.sessionNo = "changed"

	if moduleReg.session.sessionNo != ses.sessionNo {
		t.Error("no reference of session is used")
	}

}
