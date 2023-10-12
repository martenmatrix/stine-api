package stineapi

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"regexp"
)

/*
ModuleRegistration represents a registration for a module on the STiNE platform.
TanRequired is set to true, if the registration process needs action from the user because an iTAN is required.
TanStartsWith represents the two starting numbers of the required iTAN with a leading zero.
*/
type ModuleRegistration struct {
	initialLink   string
	session       *Session
	TanRequired   bool
	TanStartsWith string
}

func (session *Session) replaceSessionNumber(registrationLink string) string {
	reg := regexp.MustCompile("ARGUMENTS=-N\\d{15}")
	return reg.ReplaceAllString(registrationLink, "ARGUMENTS=-N"+session.sessionNo)
}

func (session *Session) getRegistrationId(registrationLink string) (string, error) {
	res, _ := session.client.Get(registrationLink)
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", err
	}

	regId, onPage := doc.Find(`input[name="rgtr_id"]`).First().Attr("value")
	if !onPage {
		return "", errors.New("unable to find registration id in response")
	}

	return regId, nil
}

func (session *Session) CreateModuleRegistration(registrationLink string) (ModuleRegistration, error) {
	url := session.replaceSessionNumber(registrationLink)
	regId, regIdErr := session.getRegistrationId(url)
	if regIdErr != nil {
		return ModuleRegistration{}, regIdErr
	}
	fmt.Println(regId)
	return ModuleRegistration{}, nil
}
