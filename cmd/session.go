package stineapi

import (
	"errors"
	"github.com/martenmatrix/stine-api/cmd/internal/auth"
	"github.com/martenmatrix/stine-api/cmd/internal/language"
	"github.com/martenmatrix/stine-api/cmd/internal/sessionNo"
	"github.com/martenmatrix/stine-api/cmd/internal/stineURL"
	"net/http"
	"net/url"
)

// Session represent a STiNE session. Think of it like an isolated tab with STiNE open.
type Session struct {
	Client    *http.Client // Client is an HTTP client, which is authenticated on STiNE, if Login was successful
	SessionNo string       // Identifier for the current session provided by STiNE, could be unique, empty string prior to successful Login
}

// NewSession creates a new [Session] and returns it.
func NewSession() Session {
	return Session{
		Client: auth.GetClient(),
	}
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

	// cnsc cookie is returned malformatted, set manually on Client
	cnscCookie, cookieErr := auth.GetMalformattedCnscCookie(res)
	if cookieErr != nil {
		return cookieErr
	}
	authUrl, authUrlErr := url.Parse(stineURL.Url + "/scripts")
	if authUrlErr != nil {
		return authUrlErr
	}
	session.Client.Jar.SetCookies(authUrl, []*http.Cookie{cnscCookie})

	// http library does not follow "Refresh"-Header, not in http specification
	session.SessionNo = sessionNo.Get(res.Header.Get("Refresh"))

	return nil
}

/*
Login authenticates a session on the STiNE website. If no error is returned, the user is logged in.
*/
func (session *Session) Login(username string, password string) error {
	linkToAuthForm, linkToAuthFormErr := auth.GetLinkToAuthForm(auth.StartPage, session.Client)
	if linkToAuthFormErr != nil {
		return linkToAuthFormErr
	}

	// creates inital antiforgery cookie in jar
	authFormRes, authFormResErr := session.Client.Get(linkToAuthForm)
	if authFormResErr != nil {
		return authFormResErr
	}
	defer authFormRes.Body.Close()

	authToken, authTokenErr := auth.GetAuthenticationToken(authFormRes)
	if authTokenErr != nil {
		return authTokenErr
	}
	returnURL, returnURLErr := auth.GetReturnURL(authFormRes)
	if returnURLErr != nil {
		return returnURLErr
	}

	makeSessionError := session.makeSession(returnURL, username, password, authToken, auth.AuthenticationForm)
	if makeSessionError != nil {
		return makeSessionError
	}

	return nil
}

/*
GetCategories returns the [moduleGetter.Category] with modules and nested categories the user can register for.

The depth indicates how deep different categories are nested within a category - starting at 0, which returns the initial page.
*/
func (session *Session) GetCategories(depth int) (Category, error) {
	registrationURL := sessionNo.Refresh("https://stine.uni-hamburg.de/scripts/mgrqispi.dll?APPNAME=CampusNet&PRGNAME=REGISTRATION&ARGUMENTS=-N000000000000000", session.SessionNo)
	initialCategory, err := getAvailableModules(depth, registrationURL, session.Client)
	if err != nil {
		return Category{}, err
	}

	return initialCategory, nil
}

/*
RegisterForModule registers the current authenticated user for the passed [moduleGetter.Module]. A [moduleRegisterer.ModuleRegistration] will be returned, which provides various functions for the registration.
*/
func (session *Session) RegisterForModule(module Module) *ModuleRegistration {
	return CreateModuleRegistration(module.RegistrationLink, session.SessionNo, session.Client)
}

/*
ChangeLanguage changes the language on the STiNE website for the current authenticated user.
The language parameter accepts the following values:

"en" - english

"de" - german
*/
func (session *Session) ChangeLanguage(newLanguage string) error {
	if newLanguage == "en" {
		err := language.ChangeToEnglish(session.Client, session.SessionNo)
		if err != nil {
			return err
		}
	} else if newLanguage == "de" {
		err := language.ChangeToGerman(session.Client, session.SessionNo)
		if err != nil {
			return err
		}
	} else {
		return errors.New("ChangeLanguage only accepts \"de\" or \"en\" as values")
	}
	return nil
}
