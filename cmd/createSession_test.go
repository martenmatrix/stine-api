package stineapi

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestGetLoginHrefValue(t *testing.T) {
	fakeId := "aWonderfulId"
	fakeElement := fmt.Sprintf("<a id='logIn_btn' class='img img_arrowSubmit' title='Anmelden' href='%s'>Anmelden</a>", fakeId)
	fakeResponse := &http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewBufferString(fakeElement)),
	}

	auth, err := getLoginHrefValue(fakeResponse)

	if auth != fakeId {
		t.Errorf("WANT: %s, GOT: %s", fakeId, auth)
	}
	if err != nil {
		t.Errorf("ERROR: %s", err)
	}
}

func TestGetLinkToAuthForm(t *testing.T) {
	session := NewSession()
	fakeId := "aWonderfulId"

	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<a id='logIn_btn' class='img img_arrowSubmit' title='Anmelden' href='aWonderfulId'>Anmelden</a>"))
	}),
	)
	defer fakeServer.Close()

	auth, err := session.getLinkToAuthForm(fakeServer.URL)

	if auth != fakeId {
		t.Errorf("WANT: %s, GOT: %s", fakeId, auth)
	}
	if err != nil {
		t.Errorf("ERROR: %s", err)
	}
}

func TestGetAuthenticationToken(t *testing.T) {
	fakeId := "aTestingId"
	fakeElement := fmt.Sprintf(`<input name="__RequestVerificationToken" type="hidden" value="%s"></input>`, fakeId)
	fakeResponse := &http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewBufferString(fakeElement)),
	}

	authToken, err := getAuthenticationToken(fakeResponse)
	if authToken != fakeId {
		t.Errorf("WANT: %s, GOT: %s", fakeId, authToken)
	}
	if err != nil {
		t.Errorf("ERROR: %s", err)
	}
}

func TestGetReturnURL(t *testing.T) {
	fakeRequest := &http.Request{
		URL: &url.URL{
			Host:     "cndsf.ad.uni-hamburg.de",
			RawQuery: "ReturnUrl=%2FIdentityServer%2Fconnect%2Fauthorize%2Fcallback%3Fclient_id%3D",
		},
	}
	fakeResponse := &http.Response{
		StatusCode: 200,
		Request:    fakeRequest,
	}

	returnURL, err := getReturnURL(fakeResponse)

	if returnURL != "/IdentityServer/connect/authorize/callback?client_id=" {
		t.Errorf("WANT: /IdentityServer/connect/authorize/callback?client_id=, GOT: %s", returnURL)
	}
	if err != nil {
		t.Errorf("ERROR: %s", err)
	}
}

func TestGetMalformattedCNSCCookie(t *testing.T) {
	fakeResponse := &http.Response{
		StatusCode: 200,
		Header: http.Header{
			// cookie needs to be malformatted
			"Set-Cookie": {"cnsc =DWFWDF; HttpOnly"},
		},
	}

	cnscCookie := getMalformattedCnscCookie(fakeResponse)

	if cnscCookie.Value != "DWFWDF" {
		t.Errorf("Cookies value differs from response value")
	}
	if cnscCookie.Name != "cnsc" {
		t.Errorf("Cookies name differs from response name")
	}
}

func TestGetSessionNo(t *testing.T) {
	sessionURL := "https://stine.uni-hamburg.de/scripts/mgrqispi.dll?APPNAME=CampusNet&PRGNAME=MLSSTART&ARGUMENTS=-N899462345432351,-N000266,"
	sessionNo := getSessionNo(sessionURL)

	if sessionNo != "899462345432351" {
		t.Errorf("WANT: 899462345432351, GOT: %s", sessionNo)
	}
}

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
	if session.sessionNo != "899462345432351" {
		t.Errorf("session number was not correctly set on client")
	}
	if valuesPasedCorrectly == false {
		t.Errorf("Did not receive expected form query input")
	}
}
