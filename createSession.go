package stineapi

import (
	"errors"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"net/url"
	"strings"
)

type Session struct {
	client    *http.Client
	sessionNo string
}

func NewSession() Session {
	return Session{
		client: getClient(),
	}
}

func (session *Session) getSTINEAuthURL() (string, error) {
	reqURL := "https://www.stine.uni-hamburg.de/scripts/mgrqispi.dll?APPNAME=CampusNet&PRGNAME=EXTERNALPAGES&ARGUMENTS=-N000000000000001,-N000265,-Astartseite"
	resp, err := session.client.Get(reqURL)
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
	cookieWithoutAttributes := cookieWithoutName[:indexOfFirstSemicolon]

	return &http.Cookie{
		Name:     "cnsc",
		Value:    cookieWithoutAttributes,
		Domain:   "stine.uni-hamburg.de",
		Path:     "/scripts",
		HttpOnly: true,
	}
}

// stine url are from the following format
// dispatcher + "?APPNAME=" + applicationName + "&PRGNAME=" + programName + "&ARGUMENTS=-N" + sessionNo + ",-N" + menuId  + temp_args
func getSessionNo(urlStr string) string {
	_, path, _ := strings.Cut(urlStr, "ARGUMENTS=-N")
	return path[:15]
}

// creates idsrv, idsrv.session and cnsc cookie in jar
// the cnsc cookie needs to be added manually to the jar because the server sends it malformatted
func (session *Session) makeSession(returnURL string, username string, password string, authToken string) error {
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
	res, resErr := session.client.PostForm(reqURL, formQuery)
	if resErr != nil {
		return resErr
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return errors.New("authentication with username/password failed")
	}

	// cnsc cookie is retunred malformatted, set manually on client
	cnscCookie := getMalformattedCnscCookie(res)
	stineURL, stineURLErr := url.Parse("https://stine.uni-hamburg.de/scripts")
	if stineURLErr != nil {
		return stineURLErr
	}
	session.client.Jar.SetCookies(stineURL, []*http.Cookie{cnscCookie})

	// http library does not follow "Refresh"-Header, not in http specification
	session.sessionNo = getSessionNo(res.Header.Get("Refresh"))

	return nil
}

func (session *Session) Login(username string, password string) error {
	authURL, authURLErr := session.getSTINEAuthURL()
	if authURLErr != nil {
		return authURLErr
	}

	// creates inital antiforgery cookie in jar
	authPageRes, authPageResErr := session.client.Get(authURL)
	if authPageResErr != nil {
		return authPageResErr
	}
	defer authPageRes.Body.Close()

	authToken, authTokenErr := getAuthenticationToken(authPageRes)
	if authTokenErr != nil {
		return authTokenErr
	}
	returnURL, returnURLErr := getReturnURL(authPageRes)
	if returnURLErr != nil {
		return returnURLErr
	}

	makeSessionError := session.makeSession(returnURL, username, password, authToken)
	if makeSessionError != nil {
		return makeSessionError
	}

	return nil
}
