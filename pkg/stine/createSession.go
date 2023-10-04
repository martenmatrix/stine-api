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

func getIdsrvCookies(client *http.Client, returnURL string, username string, password string, authToken string, antiForgeryCookie *http.Cookie) (idsrvCookies, error) {
	reqURL := "https://cndsf.ad.uni-hamburg.de/IdentityServer/Account/Login"
	formQuery := url.Values{
		"ReturnUrl":                  {returnURL},
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

func getRedirectFromRefreshHeader(header http.Header) string {
	refreshHeader := header.Get("Refresh")
	indexOfFirstEquals := strings.IndexByte(refreshHeader, '=')
	redirectPath := refreshHeader[indexOfFirstEquals+1:]
	redirectURL := fmt.Sprintf("https://stine.uni-hamburg.de%s", redirectPath)
	return redirectURL
}

func createHomepageRedirectRequest(authURL string, sessionCookies idsrvCookies, antiForgeryCookie *http.Cookie) (*http.Request, error) {
	req, reqErr := http.NewRequest("GET", authURL, nil)
	if reqErr != nil {
		return nil, reqErr
	}
	req.AddCookie(antiForgeryCookie)
	req.AddCookie(sessionCookies.idsrv)
	req.AddCookie(sessionCookies.idsrvSession)
	return req, nil
}

func getMalformattedCnscCookie(homepageRedirectRes *http.Response) *http.Cookie {
	setCookieHeader := homepageRedirectRes.Header.Get("Set-Cookie")
	indexOfFirstEquals := strings.IndexByte(setCookieHeader, '=')
	cookieWithoutName := setCookieHeader[indexOfFirstEquals+1:]
	indexOfFirstSemicolon := strings.IndexByte(cookieWithoutName, ';')
	cookieWithoutAttributes := cookieWithoutName[:indexOfFirstSemicolon-1]

	return &http.Cookie{
		Name:     "cnsc",
		Value:    cookieWithoutAttributes,
		HttpOnly: true,
	}
}

type name struct {
	name    string
	surname string
}

func GetSession(username string, password string) (Session, error) {
	client := getClient()

	authURL, authURLErr := getSTINEAuthURL(client)
	if authURLErr != nil {
		return Session{}, authURLErr
	}

	authPageRes, authPageResErr := client.Get(authURL)
	if authPageResErr != nil {
		return Session{}, authPageResErr
	}
	defer authPageRes.Body.Close()

	antiForgeryCookie := getAntiforgeryCookie(authPageRes)
	authToken, authTokenErr := getAuthenticationToken(authPageRes)
	returnURL := authPageRes.Request.URL.RawQuery
	if authTokenErr != nil {
		return Session{}, authTokenErr
	}

	idsrvCookies, idsrvError := getIdsrvCookies(client, returnURL, username, password, authToken, antiForgeryCookie)
	if idsrvError != nil {
		return Session{}, idsrvError
	}

	stineAuthURL, stineAuthURLErr := getSTINEAuthURL(client)
	if stineAuthURLErr != nil {
		return Session{}, stineAuthURLErr
	}

	homepageRedirectReq, homepageRedirectReqErr := createHomepageRedirectRequest(stineAuthURL, idsrvCookies, antiForgeryCookie)
	if homepageRedirectReqErr != nil {
		return Session{}, homepageRedirectReqErr
	}
	homepageRedirectRes, homepageRedirectResErr := client.Do(homepageRedirectReq)
	if homepageRedirectResErr != nil {
		return Session{}, homepageRedirectResErr
	}
	homepageURL := getRedirectFromRefreshHeader(homepageRedirectRes.Header)
	cnscCookie := getMalformattedCnscCookie(homepageRedirectRes)
	fmt.Println(homepageURL, cnscCookie)

	return Session{}, nil
}
