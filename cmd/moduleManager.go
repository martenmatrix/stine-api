package stineapi

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"regexp"
)

type ModuleRegistration struct {
	TanRequired   bool
	TanStartsWith string
	SetTan        func(tan string)
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
