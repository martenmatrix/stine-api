package moduleRegisterer

import (
	"bytes"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOniTANPage(t *testing.T) {
	fakeRes1, err1 := goquery.NewDocumentFromReader(io.NopCloser(bytes.NewBufferString("<html><body></body></html>")))
	if err1 != nil {
		t.Errorf(err1.Error())
	}

	res1 := oniTANPage(fakeRes1)

	if res1 != false {
		t.Error("Expected: false, Received: true")
	}

	fakeRes2, err2 := goquery.NewDocumentFromReader(io.NopCloser(bytes.NewBufferString("<html><body><span class=\"itan\"</body></html>")))
	if err2 != nil {
		t.Errorf(err2.Error())
	}
	res2 := oniTANPage(fakeRes2)

	if res2 != true {
		t.Error("Expected: true, Received: false")
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

func TestCheckForTanError(t *testing.T) {
	fakeResponse := &http.Response{
		Body: io.NopCloser(bytes.NewBufferString(`<span class="error">a custom error msg<span>`)),
	}
	err := checkForTANError(fakeResponse)

	if err == nil {
		t.Error("itan was sent not successfully, however no error is returned")
	}

	if !strings.Contains(err.Error(), "a custom error msg") {
		t.Error("err msg returned by stine is not contained in returned err")
	}
}

func TestSendTan(t *testing.T) {
	var valuesPassedCorrectly bool
	tan := &TanRequired{
		registration: &ModuleRegistration{
			registrationId: "23233",
			client:         &http.Client{},
			sessionNumber:  "324324",
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
			r.Form.Get("sessionno") == tan.registration.sessionNumber &&
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
