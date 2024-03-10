package moduleRegisterer

import (
	"bytes"
	"fmt"
	"github.com/PuerkitoBio/goquery"
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

	regId, err := GetRegistrationId(&http.Client{}, fakeServer.URL)

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

	rbCode, rbErr := GetRBCode(rbCodeRes)

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

	_, err := DoExamRegistrationRequest(&http.Client{}, formRequestMock.URL, "RBCODE23244", "222", "333", "444", 1)

	if err != nil {
		t.Errorf(err.Error())
	}
}
