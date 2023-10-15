package stineapi

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateModuleRegistration(t *testing.T) {
	fakeRegistrationLink := "https://stine.uni-hamburg.de/scripts/mgrqispi.dll?APPNAME=CampusNet&PRGNAME=REGCOURSEMOD&ARGUMENTS=-N232343443351119,-N343449,-N343424234011169,-ADOFF,-N343434342285453,-N344343434341730,-N0,-N0,-N0,-AN,-N0"
	ses := NewSession()
	moduleReg := ses.CreateModuleRegistration(fakeRegistrationLink)
	if moduleReg.registrationLink != fakeRegistrationLink {
		t.Error("registration link is not set on object")
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

func TestReplaceSessionNumber(t *testing.T) {
	fakeRegistrationLink := "https://stine.uni-hamburg.de/scripts/mgrqispi.dll?APPNAME=CampusNet&PRGNAME=REGCOURSEMOD&ARGUMENTS=-N232343443351119,-N343449,-N343424234011169,-ADOFF,-N343434342285453,-N344343434341730,-N0,-N0,-N0,-AN,-N0"
	newSessionNo := "232343443351119"
	ses := NewSession()
	ses.sessionNo = newSessionNo
	moduleReg := ses.CreateModuleRegistration(fakeRegistrationLink)

	moduleReg.refreshSessionNumber()

	if moduleReg.registrationLink != "https://stine.uni-hamburg.de/scripts/mgrqispi.dll?APPNAME=CampusNet&PRGNAME=REGCOURSEMOD&ARGUMENTS=-N"+newSessionNo+",-N343449,-N343424234011169,-ADOFF,-N343434342285453,-N344343434341730,-N0,-N0,-N0,-AN,-N0" {
		t.Error("session number is not being replaced in link")
	}
}

func TestGetRegistrationId(t *testing.T) {
	fakeRegistrationId := "2132134"
	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<input name="rgtr_id" value="` + fakeRegistrationId + `"/>`))
	}),
	)
	ses := NewSession()
	modReg := ses.CreateModuleRegistration("https://www.example.org")
	modReg.registrationLink = fakeServer.URL

	err := modReg.getRegistrationId()

	if err != nil {
		t.Error(err)
	}
	if modReg.registrationId != fakeRegistrationId {
		t.Error(fmt.Sprintf("EXPECTED: %s, RECEIVED: %s", fakeRegistrationId, modReg.registrationId))
	}
}

func TestSetExamDate(t *testing.T) {
	ses := NewSession()
	moduleReg := ses.CreateModuleRegistration("https://www.example.org")

	if moduleReg.examDate != 0 {
		t.Error("default value for examDate should be 0")
	}

	moduleReg.SetExamDate(1)

	if moduleReg.examDate != 1 {
		t.Error("unable to change exam date")
	}

	moduleReg.SetExamDate(3)

	if moduleReg.examDate == 3 {
		t.Error("able to pass invalid arguments (>2)")
	}

	moduleReg.SetExamDate(-1)

	if moduleReg.examDate == -1 {
		t.Error("able to pass invalid arguments (<0)")
	}
}

func TestDoRegistrationRequest(t *testing.T) {
	var valuesPassedCorrectly bool
	sessionNo := "1234"
	menuId := "54321"
	rgtrId := "232423"

	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			t.Errorf("ERROR: %s", err)
		}
		valuesPassedCorrectly = r.Form.Get("Next") == "Weiter" &&
			r.Form.Get("APPNAME") == "CAMPUSNET" &&
			r.Form.Get("PRGNAME") == "SAVEREGISTRATIONDETAILS" &&
			r.Form.Get("ARGUMENTS") == "sessionno,menuid,rgtr_id" &&
			r.Form.Get("sessionno") == sessionNo &&
			r.Form.Get("menuid") == menuId &&
			r.Form.Get("rgtr_id") == rgtrId
	}),
	)
	defer fakeServer.Close()

	ses := NewSession()
	reg := ses.CreateModuleRegistration("https://stine.uni-hamburg.de/")
	reg.session.sessionNo = sessionNo
	reg.menuId = menuId
	reg.registrationId = rgtrId
	res, err := reg.doRegistrationRequest(fakeServer.URL)
	defer res.Body.Close()

	if err != nil {
		t.Errorf(err.Error())
	}

	if valuesPassedCorrectly == false {
		t.Error("form request is not formatted correctly")
	}
}
