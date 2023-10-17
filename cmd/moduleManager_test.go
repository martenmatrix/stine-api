package stineapi

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestOniTANPage(t *testing.T) {
	fakeRes1 := &http.Response{
		Body: ioutil.NopCloser(bytes.NewBufferString("<html><body></body></html>")),
	}
	res1 := oniTANPage(fakeRes1)

	if res1 != false {
		t.Error("Expected: false, Received: true")
	}

	fakeRes2 := &http.Response{
		Body: ioutil.NopCloser(bytes.NewBufferString("<html><body><!-- CONFIRM AND TAN INPUT --><span></span></body></html>")),
	}
	res2 := oniTANPage(fakeRes2)

	if res2 != true {
		t.Error("Expected: true, Received: false")
	}
}

func TestGetTanRequiredStruct(t *testing.T) {
	fakeRes := &http.Response{
		Body: ioutil.NopCloser(bytes.NewBufferString(`<span class="itan"> 40</span>`)),
	}

	ses := NewSession()
	modReg := ses.CreateModuleRegistration("x")
	tan, err := modReg.getTanRequiredStruct(fakeRes)

	if err != nil {
		t.Errorf(err.Error())
	}

	if tan.TanStartsWith != "040" {
		t.Error(fmt.Sprintf("EXPECTED: %s, RECEIVED: %s", "040", tan.TanStartsWith))
	}

	if tan.registration.registrationLink != "x" {
		t.Error("registration struct is not correctly copied to tanrequired struct")
	}
}

func TestSendTan(t *testing.T) {
	var valuesPassedCorrectly bool
	tan := &TanRequired{
		registration: &ModuleRegistration{
			registrationId: "23233",
			session: &Session{
				Client:    &http.Client{},
				sessionNo: "324324",
			},
		},
	}

	formRequestMock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			t.Errorf("ERROR: %s", err)
		}
		valuesPassedCorrectly = r.Form.Get("campusnetsubmit") == "" &&
			r.Form.Get("tan_code") == "23" &&
			r.Form.Get("APPNAME") == "CampusNet" &&
			r.Form.Get("PRGNAME") == "SAVEREGISTRATION" &&
			r.Form.Get("ARGUMENTS") == "sessionno,menuid,rgtr_id,mode,timetable_id,location_id" &&
			r.Form.Get("sessionno") == tan.registration.session.sessionNo &&
			r.Form.Get("rgtr_id") == tan.registration.registrationId &&
			r.Form.Get("mode") == "   0"

		if valuesPassedCorrectly != true {
			t.Error(fmt.Sprintf("form was not sent with correct attributes: %s", r.Form))
		}
	}),
	)
	defer formRequestMock.Close()

	err := tan.sendTAN(formRequestMock.URL, "23")

	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestCheckForTanError(t *testing.T) {
	fakeResponse := &http.Response{
		Body: ioutil.NopCloser(bytes.NewBufferString(`<span class="error">a custom error msg<span>`)),
	}
	err := checkForTANError(fakeResponse)

	if err == nil {
		t.Error("itan was sent not successfully, however no error is returned")
	}

	if !strings.Contains(err.Error(), "a custom error msg") {
		t.Error("err msg returned by stine is not contained in returned err")
	}
}

func TestOnSelectExamPage(t *testing.T) {
	onExamPage := &http.Response{
		Body: ioutil.NopCloser(bytes.NewBufferString(`<input name="PRGNAME" type="hidden" value="SAVEEXAMDETAILS">`)),
	}

	notOnExamPage := &http.Response{
		Body: ioutil.NopCloser(bytes.NewBufferString(`<input name="PRGNAME" type="hidden">`)),
	}

	if onSelectExamPage(onExamPage) {
		t.Error("should return true")
	}

	if onSelectExamPage(notOnExamPage) {
		t.Error("should return false")
	}
}
