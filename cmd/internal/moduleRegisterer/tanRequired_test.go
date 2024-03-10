package moduleRegisterer

import (
	"bytes"
	"fmt"
	"github.com/martenmatrix/stine-api/cmd/internal/tan"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRemoveTanPrefix(t *testing.T) {
	tanReq := &TanRequired{
		TanStartsWith: "054",
	}

	prefixLeadingZero := tan.RemoveTanPrefix("05421213", tanReq.TanStartsWith)
	if prefixLeadingZero != "21213" {
		t.Error(fmt.Sprintf("should have removed 054 from tan, received %s", prefixLeadingZero))
	}

	prefixNormal := tan.RemoveTanPrefix("4242445324", tanReq.TanStartsWith)
	if prefixNormal != "4242445324" {
		t.Error(fmt.Sprintf("nothing should have been removed from tan, however received %s", prefixNormal))
	}
}

func TestCheckForTanError(t *testing.T) {
	fakeResponse := &http.Response{
		Body: io.NopCloser(bytes.NewBufferString(`<span class="error">a custom error msg<span>`)),
	}
	err := tan.CheckForTANError(fakeResponse)

	if err == nil {
		t.Error("itan was sent not successfully, however no error is returned")
	}

	if !strings.Contains(err.Error(), "a custom error msg") {
		t.Error("err msg returned by stine is not contained in returned err")
	}
}

func TestSendTan(t *testing.T) {
	var valuesPassedCorrectly bool
	fakeTAN := &TanRequired{
		client:         &http.Client{},
		registrationId: "23233",
		sessionNo:      "324324",
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
			r.Form.Get("sessionno") == fakeTAN.sessionNo &&
			r.Form.Get("rgtr_id") == fakeTAN.registrationId &&
			r.Form.Get("mode") == "   0"

		if valuesPassedCorrectly != true {
			t.Error(fmt.Sprintf("form was not sent with correct attributes: %s", r.Form))
		}
	}),
	)
	defer formRequestMock.Close()

	err := tan.SendTAN(&http.Client{}, formRequestMock.URL, "23", fakeTAN.sessionNo, fakeTAN.registrationId)

	if err != nil {
		t.Errorf(err.Error())
	}
}
