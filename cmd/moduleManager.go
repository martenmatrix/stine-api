package main
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

func main() {

func (session *Session) CreateModuleRegistration(registrationLink string) ModuleRegistration {
	url := session.replaceSessionNumber(registrationLink)
	session.getRegistrationId(url)
	return ModuleRegistration{}
}
