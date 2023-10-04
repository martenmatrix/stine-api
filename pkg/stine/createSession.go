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

func getMalformattedCnscCookie(respWithCookie *http.Response) *http.Cookie {
	setCookieHeader := respWithCookie.Header.Get("Set-Cookie")
	indexOfFirstEquals := strings.IndexByte(setCookieHeader, '=')
	cookieWithoutName := setCookieHeader[indexOfFirstEquals+1:]
	indexOfFirstSemicolon := strings.IndexByte(cookieWithoutName, ';')
	cookieWithoutAttributes := cookieWithoutName[:indexOfFirstSemicolon-1]

	return &http.Cookie{
		Name:     "cnsc",
		Value:    cookieWithoutAttributes,
		Domain:   "stine.uni-hamburg.de",
		Path:     "/scripts",
		HttpOnly: true,
	}
}

func getRedirectFromRefreshHeader(header http.Header) string {
	refreshHeader := header.Get("Refresh")
	indexOfFirstEquals := strings.IndexByte(refreshHeader, '=')
	redirectPath := refreshHeader[indexOfFirstEquals+1:]
	redirectURL := fmt.Sprintf("https://stine.uni-hamburg.de%s", redirectPath)
	return redirectURL
}

// creates idsrv, idsrv.session and cnsc cookie in jar
// the cnsc cookie needs to be added manually to the jar because the server sends it malformatted
// returns url, which re-directs to homepage, contains sessionId and is made of the following format
// dispatcher + "?APPNAME=" + applicationName + "&PRGNAME=" + programName + "&ARGUMENTS=-N" + sessionNo + ",-N" + menuId  + temp_args
func makeSessionCookies(client *http.Client, returnURL string, username string, password string, authToken string) (string, error) {
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
	if resErr != nil {
		return "", resErr
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", errors.New("authentication with username/password failed")
	}

	cnscCookie := getMalformattedCnscCookie(res)
	stineURL, stineURLErr := url.Parse("https://stine.uni-hamburg.de/scripts")
	if stineURLErr != nil {
		return "", stineURLErr
	}
	client.Jar.SetCookies(stineURL, []*http.Cookie{cnscCookie})

	client.Jar.Cookies(stineURL)

	homepageURL := getRedirectFromRefreshHeader(res.Header)
	return homepageURL, nil
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

	homepageURL, idsrvError := makeSessionCookies(client, returnURL, username, password, authToken)
	if idsrvError != nil {
		return Session{}, idsrvError
	}

	req1, _ := http.NewRequest("GET", homepageURL, nil)
	logRequest(req1)
	res1, err2 := client.Do(req1)
	if err2 != nil {
		return Session{}, authPageResErr
	}
	defer res1.Body.Close()
	logResponse(res1)

	return Session{
		client:      client,
		homepageURL: homepageURL,
	}, nil
}
