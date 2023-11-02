package stineapi

import (
	"bytes"
	"fmt"
	"github.com/PuerkitoBio/goquery"
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
		valuesPassedCorrectly = r.Form.Get("Next") == " Weiter" &&
			r.Form.Get("APPNAME") == "CampusNet" &&
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
	fakeRes1, err1 := goquery.NewDocumentFromReader(ioutil.NopCloser(bytes.NewBufferString("<html><body></body></html>")))
	if err1 != nil {
		t.Errorf(err1.Error())
	}

	res1 := oniTANPage(fakeRes1)

	if res1 != false {
		t.Error("Expected: false, Received: true")
	}

	fakeRes2, err2 := goquery.NewDocumentFromReader(ioutil.NopCloser(bytes.NewBufferString("<html><body><span class=\"itan\"</body></html>")))
	if err2 != nil {
		t.Errorf(err2.Error())
	}
	res2 := oniTANPage(fakeRes2)

	if res2 != true {
		t.Error("Expected: true, Received: false")
	}
}

func TestGetTanRequiredStruct(t *testing.T) {
	fakeRes, err := goquery.NewDocumentFromReader(ioutil.NopCloser(bytes.NewBufferString("<span class=\"itan\"> 40</span>")))
	if err != nil {
		t.Errorf(err.Error())
	}

	ses := NewSession()
	modReg := ses.CreateModuleRegistration("x")
	tan := modReg.getTanRequiredStruct(fakeRes)

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
	onExamPage, errExamPage := goquery.NewDocumentFromReader(ioutil.NopCloser(bytes.NewBufferString(`<input name="PRGNAME" type="hidden" value="SAVEEXAMDETAILS">`)))
	if errExamPage != nil {
		t.Errorf(errExamPage.Error())
	}

	notOnExamPage, errNotExamPage := goquery.NewDocumentFromReader(ioutil.NopCloser(bytes.NewBufferString(`<input name="PRGNAME" type="hidden">`)))
	if errNotExamPage != nil {
		t.Errorf(errNotExamPage.Error())
	}

	if onSelectExamPage(onExamPage) != true {
		t.Error("should return true")
	}

	if onSelectExamPage(notOnExamPage) != false {
		t.Error("should return false")
	}
}

func TestGetRbCode(t *testing.T) {
	rbCodeRes, err := goquery.NewDocumentFromReader(ioutil.NopCloser(bytes.NewBufferString(`
		<input name="trap" class="checkBox" value=" 1">
		<input type="radio" class="checkBox" name="RB_388233088543" value=" 1">
	`)))
	if err != nil {
		t.Errorf(err.Error())
	}

	rbCode, rbErr := getRBCode(rbCodeRes)

	if rbErr != nil {
		t.Errorf(rbErr.Error())
	}

	if rbCode != "RB_388233088543" {
		t.Error(fmt.Sprintf("Expected: RB_388233088543, Received %s", rbCode))
	}
}

func TestDoExamRegistrationRequest(t *testing.T) {
	var valuesPassedCorrectly bool
	ses := NewSession()
	modReg := ses.CreateModuleRegistration("")
	modReg.SetExamDate(1)

	formRequestMock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			t.Errorf("ERROR: %s", err)
		}
		valuesPassedCorrectly = r.Form.Get("Next") == " Next" &&
			r.Form.Get("RBCODE23244") == " 2" &&
			r.Form.Get("APPNAME") == "CAMPUSNET" &&
			r.Form.Get("PRGNAME") == "SAVEEXAMDETAILS" &&
			r.Form.Get("ARGUMENTS") == "sessionno,menuid,rgtr_id,mode" &&
			r.Form.Get("sessionno") == modReg.session.sessionNo &&
			r.Form.Get("menuid") == modReg.menuId &&
			r.Form.Get("rgtr_id") == modReg.registrationId &&
			r.Form.Get("mode") == "0001"

		if valuesPassedCorrectly != true {
			t.Error(fmt.Sprintf("form was not sent with correct attributes: %s", r.Form))
		}
	}),
	)
	defer formRequestMock.Close()

	_, err := modReg.doExamRegistrationRequest(formRequestMock.URL, "RBCODE23244")

	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestRemoveTanPrefix(t *testing.T) {
	tanReq := &TanRequired{
		TanStartsWith: "054",
	}

	prefixLeadingZero := tanReq.removeTanPrefix("05421213")
	if prefixLeadingZero != "21213" {
		t.Error(fmt.Sprintf("should have removed 054 from tan, received %s", prefixLeadingZero))
	}

	prefixNormal := tanReq.removeTanPrefix("4242445324")
	if prefixNormal != "4242445324" {
		t.Error(fmt.Sprintf("nothing should have been removed from tan, however received %s", prefixNormal))
	}
}

func TestRegister(t *testing.T) {
	// TODO maybe one big test is easier for refactoring instead of tests for every single function
	var requestCounter int
	fakeRegistrationId := "2132134"
	rbCode := "RB_2302138093248321094"

	fakeServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		requestCounter++

		switch {
		case requestCounter == 1:
			// get registration id request
			_, err := writer.Write([]byte(`<input name="rgtr_id" value="` + fakeRegistrationId + `"/>`))
			if err != nil {
				t.Errorf(err.Error())
			}
		case requestCounter == 2:
			// check if registration id was parsed correctly from previous mocked response
			errForm := request.ParseForm()
			if errForm != nil {
				t.Errorf("ERROR: %s", errForm)
			}
			rgtrId := request.Form.Get("rgtr_id")
			if rgtrId != fakeRegistrationId {
				t.Error(fmt.Sprintf("expected %s as form parameter, received %s", fakeRegistrationId, rgtrId))
			}

			// module registration request, pretending were on exam page
			_, err := writer.Write([]byte(`
				<input name="PRGNAME" type="hidden" value="SAVEEXAMDETAILS">
				<input />
				<input type="radio" class="checkBox" name="` + rbCode + `" value="  1" />
			`))
			if err != nil {
				t.Errorf(err.Error())
			}
		case requestCounter == 3:
			// check if rb code was parsed correctly from previous mocked response
			errForm := request.ParseForm()
			if errForm != nil {
				t.Errorf("ERROR: %s", errForm)
			}
			rbCodeValue := request.Form.Get(rbCode)
			if rbCodeValue != " 1" {
				t.Error(fmt.Sprintf("tried to retrieve the value of %s and got %s", rbCode, rbCodeValue))
			}

			// tan request, pretending we need an itan auth
			_, err := writer.Write([]byte(`
				<span class="itan"> 54</span>
			`))
			if err != nil {
				t.Errorf(err.Error())
			}
		case requestCounter == 4:
			// check if itan was sent correctly
			errForm := request.ParseForm()
			if errForm != nil {
				t.Errorf("ERROR: %s", errForm)
			}
			iTanValue := request.Form.Get("tan_code")
			if iTanValue != "3423" {
				t.Error(fmt.Sprintf("itan value was not parsed correctly, Expected: %s, Received: %s", "3423", iTanValue))
			}
		}
	}))

	ses := NewSession()
	modReg := ses.CreateModuleRegistration(fakeServer.URL)
	tanReq, err := modReg.Register()
	if err != nil {
		t.Errorf(err.Error())
	}

	if tanReq == nil {
		t.Error("an itan is required, however no tanrequired object was returned")
	}

	// prefix of tan is 054
	// supply tan, prefix should be removed
	tanReqErr := tanReq.SetTan("0543423")
	if tanReqErr != nil {
		t.Errorf(tanReqErr.Error())
	}

	if requestCounter != 4 {
		t.Error(fmt.Sprintf("expected 4 requests, however received %d", requestCounter))
	}
}
