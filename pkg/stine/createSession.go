package stine

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"net/url"
	"strings"
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

func makeAntiforgeryCookieAndGetAuthToken(client *http.Client) (string, error) {
	reqURL := "https://cndsf.ad.uni-hamburg.de/IdentityServer/Account/Login?ReturnUrl=%2FIdentityServer%2Fconnect%2Fauthorize%2Fcallback%3Fclient_id%3DClassicWeb%26scope%3Dopenid%2520DSF%26response_mode%3Dquery%26response_type%3Dcode%26nonce%3DkQEKs7lCwN2CEXvCDeD1Zw%253D%253D%26redirect_uri%3Dhttps%253A%252F%252Fstine.uni-hamburg.de%252Fscripts%252Fmgrqispi.dll%253FAPPNAME%253DCampusNet%2526PRGNAME%253DLOGINCHECK%2526ARGUMENTS%253D-N000000000000001,ids_mode%2526ids_mode%253DY"
	resp, getError := client.Get(reqURL)
	if getError != nil {
		return "", getError
	}
	defer resp.Body.Close()
	authToken, tokenError := getAuthenticationToken(resp)
	if tokenError != nil {
		return "", tokenError
	}
	for _, cookie := range resp.Cookies() {
		fmt.Println(cookie)
	}
	return authToken, nil
}

func makeIDSRVCookies(client *http.Client, username string, password string, authToken string) error {
	reqURL := "https://cndsf.ad.uni-hamburg.de/IdentityServer/Account/Login"
	formData := url.Values{
		"ReturnUrl":                  {},
		"CancelUrl":                  {},
		"Username":                   {username},
		"Password":                   {password},
		"RememberLogin":              {"true"},
		"button":                     {"login"},
		"__RequestVerificationToken": {authToken},
	}
	resp, err := client.PostForm(reqURL, formData)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.New("authentication with username/password failed")
	}
	// unable to log cookies with resp.Cookies(), however they are set in the cookie jar (possibly bug)
	return nil
}

func getSTINEAuthURL(client *http.Client) (string, error) {
	reqURL := "https://www.stine.uni-hamburg.de/scripts/mgrqispi.dll?APPNAME=CampusNet&PRGNAME=EXTERNALPAGES&ARGUMENTS=-N000000000000001,-N000265,-Astartseite"
	resp, err := client.Get(reqURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", err
	}

	authURL, onPage := doc.Find("#logIn_btn").First().Attr("href")
	if !onPage {
		return "", errors.New("unable to find login button on STiNE page")
	}

	return authURL, nil
}

func makeCNSCCookieAndGetLocationURL(client *http.Client) (string, error) {
	authURL, authURLErr := getSTINEAuthURL(client)
	if authURLErr != nil {
		return "", authURLErr
	}
	resp, getErr := client.Get(authURL)
	if getErr != nil {
		return "", getErr
	}
	defer resp.Body.Close()
	refreshHeader := resp.Header.Get("Refresh")
	indexOfFirstEquals := strings.IndexByte(refreshHeader, '=')
	path := refreshHeader[indexOfFirstEquals+1:]
	homePageURL := fmt.Sprintf("https://stine.uni-hamburg.de%s", path)
	return homePageURL, nil
}

func GetSession(username string, password string) error {
	client := getClient()
	authToken, antiForgeryAuthError := makeAntiforgeryCookieAndGetAuthToken(client)
	if antiForgeryAuthError != nil {
		return antiForgeryAuthError
	}
	idsrvError := makeIDSRVCookies(client, username, password, authToken)
	if idsrvError != nil {
		return idsrvError
	}
	homepageURL, csncErr := makeCNSCCookieAndGetLocationURL(client)
	if csncErr != nil {
		return csncErr
	}
	fmt.Println(homepageURL)
	return nil
}
