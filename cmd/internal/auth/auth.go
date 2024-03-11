package auth

import (
	"errors"
	"github.com/PuerkitoBio/goquery"
	"github.com/martenmatrix/stine-api/cmd/internal/stineURL"
	"net/http"
	"net/url"
	"strings"
)

const (
	StartPage          = stineURL.Url + "/scripts/mgrqispi.dll?APPNAME=CampusNet&PRGNAME=EXTERNALPAGES&ARGUMENTS=-N000000000000001,-N000265,-Astartseite"
	AuthenticationForm = "https://cndsf.ad.uni-hamburg.de/IdentityServer/Account/Login"
)

func getLoginHrefValue(resp *http.Response) (string, error) {
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

func GetLinkToAuthForm(startPageURL string, client *http.Client) (string, error) {
	resp, err := client.Get(startPageURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	authURL, authURLError := getLoginHrefValue(resp)
	if authURLError != nil {
		return "", authURLError
	}
	return authURL, nil
}

func GetAuthenticationToken(authPageRes *http.Response) (string, error) {
	doc, err := goquery.NewDocumentFromReader(authPageRes.Body)
	if err != nil {
		return "", err
	}

	selection := doc.Find("input[name='__RequestVerificationToken']").First()
	authToken, onPage := selection.Attr("value")

	if !onPage {
		return "", errors.New("unable to find authentication token")
	}

	return authToken, nil
}

func GetReturnURL(authPageRes *http.Response) (string, error) {
	returnUrl := authPageRes.Request.URL.RawQuery
	indexOfFirstEquals := strings.IndexByte(returnUrl, '=')
	returnURLWithoutName := returnUrl[indexOfFirstEquals+1:]
	decodedStr, err := url.QueryUnescape(returnURLWithoutName)
	if err != nil {
		return "", err
	}
	return decodedStr, nil
}

// GetMalformattedCnscCookie extracts a cookie, which is sent malformed by the STiNE server, which the Go Client would not parse
func GetMalformattedCnscCookie(respWithCookie *http.Response) (*http.Cookie, error) {
	setCookieHeader := respWithCookie.Header.Get("Set-Cookie")
	// no auth cookie response from server => could be wrong password
	if setCookieHeader == "" {
		return nil, errors.New("auth failed, re-check user credentials")
	}

	keyValue := strings.Split(setCookieHeader, "=")
	cookieWithoutName := keyValue[1]
	cookieValueAndAttributes := strings.Split(cookieWithoutName, ";")
	cookieValue := cookieValueAndAttributes[0]

	return &http.Cookie{
		Name:     "cnsc",
		Value:    cookieValue,
		Domain:   "stine.uni-hamburg.de",
		Path:     "/scripts",
		HttpOnly: true,
	}, nil
}
