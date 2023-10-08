package stineapi

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
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
