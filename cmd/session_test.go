package stineapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMakeSession(t *testing.T) {
	session := NewSession()
	var valuesPasedCorrectly bool

	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Set-Cookie", "cnsc=DWFWDF")
		w.Header().Add("Refresh", "https://stine.uni-hamburg.de/scripts/mgrqispi.dll?APPNAME=CampusNet&PRGNAME=MLSSTART&ARGUMENTS=-N899462345432351,-N000266,")
		err := r.ParseForm()
		if err != nil {
			t.Errorf("ERROR: %s", err)
		}
		valuesPasedCorrectly = r.Form.Get("ReturnUrl") == "peter" &&
			r.Form.Get("Username") == "user" &&
			r.Form.Get("Password") == "pass" &&
			r.Form.Get("__RequestVerificationToken") == "token" &&
			r.Form.Get("button") == "login" &&
			r.Form.Get("RememberLogin") == "true"
	}),
	)
	defer fakeServer.Close()

	err := session.makeSession("peter", "user", "pass", "token", fakeServer.URL)
	if err != nil {
		t.Errorf("ERROR: %s", err)
	}
	if session.SessionNo != "899462345432351" {
		t.Errorf("session number was not correctly set on client")
	}
	if valuesPasedCorrectly == false {
		t.Errorf("Did not receive expected form query input")
	}
}
