package stineapi

import (
	"log"
	"testing"
)

func TestArgsWithoutSessionNumber(t *testing.T) {
	ses := NewSession()
	err := ses.Login("BBB8542", "Xubnyg-pohpek-fytce2")
	if err != nil {
		log.Fatal(err)
	}
	ses.CreateModuleRegistration("https://stine.uni-hamburg.de/scripts/mgrqispi.dll?APPNAME=CampusNet&PRGNAME=REGCOURSEMOD&ARGUMENTS=-N266412591351119,-N000309,-N383764423011169,-AMOFF,-N386935112285453,-N384245998551730,-N0,-N0,-N0,-AN,-N0")
}

func TestCreateModulRegistration(t *testing.T) {
	fakeRegistrationLink := "https://stine.uni-hamburg.de/scripts/mgrqispi.dll?APPNAME=CampusNet&PRGNAME=REGCOURSEMOD&ARGUMENTS=-N232343443351119,-N343449,-N343424234011169,-ADOFF,-N343434342285453,-N344343434341730,-N0,-N0,-N0,-AN,-N0"
	ses := NewSession()
	moduleReg := ses.CreateModuleRegistration(fakeRegistrationLink)
	if moduleReg.initialLink != fakeRegistrationLink {
		t.Error("registration link is not set on object")
	}
	if moduleReg.TanRequired != false || moduleReg.TanStartsWith != "" {
		t.Error("TanRequired and TanStartsWith do not have default values")
	}
}

func TestReferenceCopyOfSessionIsUsed(t *testing.T) {

}
