package stineapi

import (
	"errors"
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


func main() {

func (session *Session) CreateModuleRegistration(registrationLink string) ModuleRegistration {
	url := session.replaceSessionNumber(registrationLink)
	session.getRegistrationId(url)
	return ModuleRegistration{}
}
