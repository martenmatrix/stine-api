package auth

import (
	"bytes"
	"fmt"
	"io"
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
		Body:       io.NopCloser(bytes.NewBufferString(fakeElement)),
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
	fakeId := "aWonderfulId"

	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<a id='logIn_btn' class='img img_arrowSubmit' title='Anmelden' href='aWonderfulId'>Anmelden</a>"))
	}),
	)
	defer fakeServer.Close()

	auth, err := GetLinkToAuthForm(fakeServer.URL, &http.Client{})

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
		Body:       io.NopCloser(bytes.NewBufferString(fakeElement)),
	}

	authToken, err := GetAuthenticationToken(fakeResponse)
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

	returnURL, err := GetReturnURL(fakeResponse)

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

	cnscCookie, err := GetMalformattedCnscCookie(fakeResponse)

	if err != nil {
		t.Errorf(err.Error())
	}

	if cnscCookie.Value != "DWFWDF" {
		t.Errorf("Cookies value differs from response value")
	}
	if cnscCookie.Name != "cnsc" {
		t.Errorf("Cookies name differs from response name")
	}

	// returns an error, if no authentication cookie was passed
	fakeResponse2 := &http.Response{
		StatusCode: 200,
		Header: http.Header{
			// cookie needs to be malformatted
			"Set-Cookie": {},
		},
	}

	_, err = GetMalformattedCnscCookie(fakeResponse2)

	if err == nil {
		t.Errorf("Function should return error, as no cnsc cookies was returned")
	}
}
