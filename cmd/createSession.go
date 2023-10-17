package stineapi

import (
	"errors"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"net/url"
	"strings"
)

const (
	startPage          = "https://www.stine.uni-hamburg.de/scripts/mgrqispi.dll?APPNAME=CampusNet&PRGNAME=EXTERNALPAGES&ARGUMENTS=-N000000000000001,-N000265,-Astartseite"
	authenticationForm = "https://cndsf.ad.uni-hamburg.de/IdentityServer/Account/Login"
)

// Session represent a STiNE session. Think of it like an isolated tab with STiNE open.
type Session struct {
	Client    *http.Client // Client is an HTTP client, which is authenticated on STiNE, after successfully executing [Login]
	sessionNo string
}

// NewSession creates a new [Session] and returns it.
func NewSession() Session {
	return Session{
		Client: getClient(),
	}
}

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

func (session *Session) getLinkToAuthForm(startPageURL string) (string, error) {
	resp, err := session.Client.Get(startPageURL)
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
	}
}

// stine urls are from the following format
// dispatcher + "?APPNAME=" + applicationName + "&PRGNAME=" + programName + "&ARGUMENTS=-N" + sessionNo + ",-N" + menuId  + temp_args
func getSessionNo(urlStr string) string {
	_, path, _ := strings.Cut(urlStr, "ARGUMENTS=-N")
	return path[:15]
}

// creates idsrv, idsrv.session and cnsc cookie in jar
// the cnsc cookie needs to be added manually to the jar because the server sends it malformatted
func (session *Session) makeSession(returnURL string, username string, password string, authToken string, authenticationFormURL string) error {
	formQuery := url.Values{
		"ReturnUrl":                  {returnURL},
		"CancelUrl":                  {},
		"Username":                   {username},
		"Password":                   {password},
		"RememberLogin":              {"true"},
		"button":                     {"login"},
		"__RequestVerificationToken": {authToken},
	}
	res, resErr := session.Client.PostForm(authenticationFormURL, formQuery)
	if resErr != nil {
		return resErr
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return errors.New("authentication with username/password failed")
	}

	// cnsc cookie is returned malformatted, set manually on client
	cnscCookie := getMalformattedCnscCookie(res)
	stineURL, stineURLErr := url.Parse("https://stine.uni-hamburg.de/scripts")
	if stineURLErr != nil {
		return stineURLErr
	}
	session.Client.Jar.SetCookies(stineURL, []*http.Cookie{cnscCookie})

	// http library does not follow "Refresh"-Header, not in http specification
	session.sessionNo = getSessionNo(res.Header.Get("Refresh"))

	return nil
}

// Login authenticates a session on the STiNE website. If no error is returned, the user is logged in. The idsvr, idsrv.session and cnsc cookie are added to a cookie jar and the session number is retrieved from the URL.
func (session *Session) Login(username string, password string) error {
	linkToAuthForm, linkToAuthFormErr := session.getLinkToAuthForm(startPage)
	if linkToAuthFormErr != nil {
		return linkToAuthFormErr
	}

	// creates inital antiforgery cookie in jar
	authFormRes, authFormResErr := session.Client.Get(linkToAuthForm)
	if authFormResErr != nil {
		return authFormResErr
	}
	defer authFormRes.Body.Close()

	authToken, authTokenErr := getAuthenticationToken(authFormRes)
	if authTokenErr != nil {
		return authTokenErr
	}
	returnURL, returnURLErr := getReturnURL(authFormRes)
	if returnURLErr != nil {
		return returnURLErr
	}

	makeSessionError := session.makeSession(returnURL, username, password, authToken, authenticationForm)
	if makeSessionError != nil {
		return makeSessionError
	}

	return nil
}
