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

func getReturnURL(authPageRes *http.Response) (string, error) {
	returnUrl := authPageRes.Request.URL.RawQuery
	indexOfFirstEquals := strings.IndexByte(returnUrl, '=')
	returnURLWithoutName := returnUrl[indexOfFirstEquals+1:]
	decodedStr, err := url.QueryUnescape(returnURLWithoutName)
	if err != nil {
		return "", err
	}
	return decodedStr, nil
}

// creates idsrv, idsrv.session and cnsc cookie in jar
func makeSessionCookies(client *http.Client, returnURL string, username string, password string, authToken string) error {
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
	res, resErr := client.PostForm(reqURL, formQuery)
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return errors.New("authentication with username/password failed")
	}

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

func getRedirectFromRefreshHeader(header http.Header) string {
	refreshHeader := header.Get("Refresh")
	indexOfFirstEquals := strings.IndexByte(refreshHeader, '=')
	redirectPath := refreshHeader[indexOfFirstEquals+1:]
	redirectURL := fmt.Sprintf("https://stine.uni-hamburg.de%s", redirectPath)
	return redirectURL
}

func createHomepageRedirectRequest(authURL string, sessionCookies string, antiForgeryCookie *http.Cookie) (*http.Request, error) {
	req, reqErr := http.NewRequest("GET", authURL, nil)
	if reqErr != nil {
		return nil, reqErr
	}
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

	// creates inital antiforgery cookie in jar
	authPageRes, authPageResErr := client.Get(authURL)
	if authPageResErr != nil {
		return Session{}, authPageResErr
	}
	defer authPageRes.Body.Close()

	authToken, authTokenErr := getAuthenticationToken(authPageRes)
	if authTokenErr != nil {
		return Session{}, authTokenErr
	}
	returnURL, returnURLErr := getReturnURL(authPageRes)
	if returnURLErr != nil {
		return Session{}, returnURLErr
	}

	idsrvError := makeSessionCookies(client, returnURL, username, password, authToken)
	if idsrvError != nil {
		return Session{}, idsrvError
	}

	homepageRedirectReq, homepageRedirectReqErr := createHomepageRedirectRequest("edddw", "f", nil)
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
