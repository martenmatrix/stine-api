package stineapi

import (
	"bytes"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/martenmatrix/stine-api/cmd/internal/moduleRegisterer"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateModuleRegistration(t *testing.T) {
	fakeRegistrationLink := "https://stine.uni-hamburg.de/scripts/mgrqispi.dll?APPNAME=CampusNet&PRGNAME=REGCOURSEMOD&ARGUMENTS=-N232343443351119,-N343449,-N343424234011169,-ADOFF,-N343434342285453,-N344343434341730,-N0,-N0,-N0,-AN,-N0"
	moduleReg := CreateModuleRegistration(fakeRegistrationLink)
	if moduleReg.registrationLink != fakeRegistrationLink {
		t.Error("registration link is not set on object")
	}
}

func TestSetExamDate(t *testing.T) {
	moduleReg := CreateModuleRegistration("https://www.example.org")

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

	reg := CreateModuleRegistration("https://stine.uni-hamburg.de/")
	reg.menuId = menuId
	reg.registrationId = rgtrId
	res, err := moduleRegisterer.DoRegistrationRequest(&http.Client{}, fakeServer.URL, sessionNo, menuId, rgtrId)
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

	modReg := CreateModuleRegistration("x")
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

	modReg := CreateModuleRegistration(fakeServer.URL)
	tanReq, err := modReg.Register(&http.Client{}, "342424")
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
