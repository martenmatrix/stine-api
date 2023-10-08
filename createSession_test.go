package stineapi

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
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
	} else if err != nil {
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
