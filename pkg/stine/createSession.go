package stine

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"net/url"
)

func getAuthenticationToken(response *http.Response) (string, error) {
	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return "", err
	}

	selection := doc.Find("input[name='__RequestVerificationToken']").First()
	authToken, onPage := selection.Attr("value")

	if !onPage {
		err = errors.New("unable to find authentication token")
		return "", err
	}

	return authToken, nil
}

func makeAntiforgeryCookie(client *http.Client) error {
	reqURL := "https://cndsf.ad.uni-hamburg.de/IdentityServer/Account/Login?ReturnUrl=%2FIdentityServer%2Fconnect%2Fauthorize%2Fcallback%3Fclient_id%3DClassicWeb%26scope%3Dopenid%2520DSF%26response_mode%3Dquery%26response_type%3Dcode%26nonce%3DkQEKs7lCwN2CEXvCDeD1Zw%253D%253D%26redirect_uri%3Dhttps%253A%252F%252Fstine.uni-hamburg.de%252Fscripts%252Fmgrqispi.dll%253FAPPNAME%253DCampusNet%2526PRGNAME%253DLOGINCHECK%2526ARGUMENTS%253D-N000000000000001,ids_mode%2526ids_mode%253DY"
	resp, getError := client.Get(reqURL)
	if getError != nil {
		return getError
	}
	defer resp.Body.Close()
	authToken, tokenError := getAuthenticationToken(resp)
	if tokenError != nil {
		return tokenError
	}
	fmt.Println(authToken)
	return nil
}

func makeIDSRVCookies(client *http.Client, username string, password string) error {
	reqURL := "https://cndsf.ad.uni-hamburg.de/IdentityServer/Account/Login?ReturnUrl=%2FIdentityServer%2Fconnect%2Fauthorize%2Fcallback%3Fclient_id%3DClassicWeb%26scope%3Dopenid%2520DSF%26response_mode%3Dquery%26response_type%3Dcode%26nonce%3DkQEKs7lCwN2CEXvCDeD1Zw%253D%253D%26redirect_uri%3Dhttps%253A%252F%252Fstine.uni-hamburg.de%252Fscripts%252Fmgrqispi.dll%253FAPPNAME%253DCampusNet%2526PRGNAME%253DLOGINCHECK%2526ARGUMENTS%253D-N000000000000001,ids_mode%2526ids_mode%253DY"
	formData := url.Values{
		"Username":      {username},
		"Password":      {password},
		"RememberLogin": {"true"},
	}
	resp, err := client.PostForm(reqURL, formData)
	if err != nil {
		return err
	}
	err = resp.Body.Close()
	if err != nil {
		return err
	}
	for _, cookie := range resp.Cookies() {
		fmt.Println(cookie)
	}
	return nil
}

func GetSession(username string, password string) error {
	client := getClient()
	antiForgeryError := makeAntiforgeryCookie(client)
	if antiForgeryError != nil {
		return antiForgeryError
	}
	idsrvError := makeIDSRVCookies(client, username, password)
	if idsrvError != nil {
		return idsrvError
	}
	return nil
}
