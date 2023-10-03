package stine

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type Session struct {
	client      *http.Client
	homepageURL string
}

func logResponse(response *http.Response) {
	resDump, err := httputil.DumpResponse(response, true)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("RESPONSE:\n%s", string(resDump))
}

func logRequest(request *http.Request) {
	reqDump, err := httputil.DumpRequest(request, true)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("REQUEST:\n%s", string(reqDump))
}

func getAntiforgeryCookie(authPageRes *http.Response) *http.Cookie {
	return authPageRes.Cookies()[0]
}

func getAuthenticationToken(authPageRes *http.Response) (string, error) {
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

type idsrvCookies struct {
	idsrv        *http.Cookie
	idsrvSession *http.Cookie
}

func getIdsrvCookies(client *http.Client, username string, password string, authToken string, antiForgeryCookie *http.Cookie) (idsrvCookies, error) {
	reqURL := "https://cndsf.ad.uni-hamburg.de/IdentityServer/Account/Login"
	formQuery := url.Values{
		"ReturnUrl":                  {},
		"CancelUrl":                  {},
		"Username":                   {username},
		"Password":                   {password},
		"RememberLogin":              {"true"},
		"button":                     {"login"},
		"__RequestVerificationToken": {authToken},
	}
	req, reqErr := http.NewRequest("POST", reqURL, strings.NewReader(formQuery.Encode()))
	if reqErr != nil {
		return idsrvCookies{}, reqErr
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(antiForgeryCookie)
	res, resErr := client.Do(req)
	if resErr != nil {
		return idsrvCookies{}, resErr
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return idsrvCookies{}, errors.New("authentication with username/password failed")
	}
	firstRedirectRes := res.Request.Response
	sessionCookies := firstRedirectRes.Cookies()

	return idsrvCookies{
		idsrv:        sessionCookies[1],
		idsrvSession: sessionCookies[0],
	}, nil
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

func getHomepageRedirect(authResp *http.Response) string {
	redirectPathHeader := authResp.Header.Get("Refresh")
	indexOfFirstEquals := strings.IndexByte(redirectPathHeader, '=')
	redirectPath := redirectPathHeader[indexOfFirstEquals+1:]
	redirectURL := fmt.Sprintf("https://stine.uni-hamburg.de%s", redirectPath)
	return redirectURL
}

func getCNSCCookie(authResp *http.Response) *http.Cookie {
	return authResp.Cookies()[0]
}

type name struct {
	name    string
	surname string
}

func GetSession(username string, password string) (Session, error) {
	client := getClient()

	authPageRes, authPageResErr := client.Get("https://cndsf.ad.uni-hamburg.de/IdentityServer/Account/Login?ReturnUrl=%2FIdentityServer%2Fconnect%2Fauthorize%2Fcallback%3Fclient_id%3DClassicWeb%26scope%3Dopenid%2520DSF%26response_mode%3Dquery%26response_type%3Dcode%26nonce%3DkQEKs7lCwN2CEXvCDeD1Zw%253D%253D%26redirect_uri%3Dhttps%253A%252F%252Fstine.uni-hamburg.de%252Fscripts%252Fmgrqispi.dll%253FAPPNAME%253DCampusNet%2526PRGNAME%253DLOGINCHECK%2526ARGUMENTS%253D-N000000000000001,ids_mode%2526ids_mode%253DY")
	if authPageResErr != nil {
		return Session{}, authPageResErr
	}

	antiForgeryCookie := getAntiforgeryCookie(authPageRes)
	authToken, authTokenErr := getAuthenticationToken(authPageRes)
	authPageRes.Body.Close()
	if authTokenErr != nil {
		return Session{}, authTokenErr
	}

	idsrvCookies, idsrvError := getIdsrvCookies(client, username, password, authToken, antiForgeryCookie)
	if idsrvError != nil {
		return Session{}, idsrvError
	}

	stineAuthURL, stineAuthURLErr := getSTINEAuthURL(client)
	if stineAuthURL != nil {
		return Session{}, stineAuthURLErr
	}

	authRes, authResErr := client.Get(stineAuthURL)
	if authResErr != nil {
		return Session{}, authResErr
	}
	defer authReq.Body.Close()
	homepageRedirect := getHomepageRedirect(authRes)
	cnscCookie := getCNSCCookie(authRes)

	fmt.Println(homepageRedirect, cnscCookie)

	return Session{
		client:      client,
		homepageURL: homepageURL,
	}, nil
}
