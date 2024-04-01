package stineapi

import (
	"bytes"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetRegistrationId(t *testing.T) {
	fakeRegistrationId := "2132134"
	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<input name="rgtr_id" value="` + fakeRegistrationId + `"/>`))
	}),
	)

	regId, err := getRegistrationId(&http.Client{}, fakeServer.URL)

	if err != nil {
		t.Error(err)
	}
	if regId != fakeRegistrationId {
		t.Error(fmt.Sprintf("EXPECTED: %s, RECEIVED: %s", fakeRegistrationId, regId))
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
			r.Form.Get("sessionno") == "222" &&
			r.Form.Get("menuid") == "333" &&
			r.Form.Get("rgtr_id") == "444" &&
			r.Form.Get("mode") == "0001"

		if valuesPassedCorrectly != true {
			t.Error(fmt.Sprintf("form was not sent with correct attributes: %s", r.Form))
		}
	}),
	)
	defer formRequestMock.Close()

	_, err := doExamRegistrationRequest(&http.Client{}, formRequestMock.URL, "RBCODE23244", "222", "333", "444", 1)

	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestCreateModuleRegistration(t *testing.T) {
	fakeRegistrationLink := "https://stine.uni-hamburg.de/scripts/mgrqispi.dll?APPNAME=CampusNet&PRGNAME=REGCOURSEMOD&ARGUMENTS=-N232343443351119,-N343449,-N343424234011169,-ADOFF,-N343434342285453,-N344343434341730,-N0,-N0,-N0,-AN,-N0"
	moduleReg := createModuleRegistration(fakeRegistrationLink, "232323", &http.Client{})
	if moduleReg.registrationLink != fakeRegistrationLink {
		t.Error("registration link is not set on object")
	}
}

func TestSetExamDate(t *testing.T) {
	moduleReg := createModuleRegistration("https://www.example.org", "232323", &http.Client{})

	if moduleReg.ExamDate != 0 {
		t.Error("default value for examDate should be 0")
	}

	moduleReg.SetExamDate(1)

	if moduleReg.ExamDate != 1 {
		t.Error("unable to change exam date")
	}

	moduleReg.SetExamDate(3)

	if moduleReg.ExamDate == 3 {
		t.Error("able to pass invalid arguments (>2)")
	}

	moduleReg.SetExamDate(-1)

	if moduleReg.ExamDate == -1 {
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

	reg := createModuleRegistration("https://stine.uni-hamburg.de/", "232323", &http.Client{})
	reg.menuId = menuId
	reg.registrationId = rgtrId
	res, err := doRegistrationRequest(&http.Client{}, fakeServer.URL, sessionNo, menuId, rgtrId)
	defer res.Body.Close()

	if err != nil {
		t.Errorf(err.Error())
	}

	if valuesPassedCorrectly == false {
		t.Error("form request is not formatted correctly")
	}
}

func TestGetTanRequiredStruct(t *testing.T) {
	fakeRes, err := goquery.NewDocumentFromReader(io.NopCloser(bytes.NewBufferString("<span class=\"itan\"> 40</span>")))
	if err != nil {
		t.Errorf(err.Error())
	}

	modReg := createModuleRegistration("x", "232323", &http.Client{})
	tan := modReg.getTanRequiredStruct(fakeRes)

	if err != nil {
		t.Errorf(err.Error())
	}

	if tan.TanStartsWith != "040" {
		t.Error(fmt.Sprintf("EXPECTED: %s, RECEIVED: %s", "040", tan.TanStartsWith))
	}

	if tan.url != "x" {
		t.Error("registration struct is not correctly copied to tanrequired struct")
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

	modReg := createModuleRegistration(fakeServer.URL, "342424", &http.Client{})
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
