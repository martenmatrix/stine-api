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
	registrationLink string
	registrationId   string
	session          *Session
	TanRequired      bool
	TanStartsWith    string
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

/*
CreateModuleRegistration creates and returns a ModuleRegistration, which provides functions to register for the specified module and its corresponding events.

It takes two arguments, the first is the link, which navigates the user to the registration process on the STiNE website.
It can be retrieved the following way from the STiNE website:
1. Navigate to STiNE and login.
2. Navigate to the module subsection, where your module is listed (e.g. for Software Development I when studying Computer Science, go to "Studying" > "Register for modules and courses" > "Compulsory Modules Informatics")
3. Your module should now be displayed with a bunch of other modules.
4. There should be a red "Register" button to the right of the module name.
5. Right-click the button and click "Copy link address", this is the registration link for the module!
If there is no "Register" button, you've either already completed the module or you've already signed up for the module.


The second argument specifies, which exam date should be selected:
1 - first exam date
2 - second exam date
3 - exam date in another semester
*/
func (session *Session) CreateModuleRegistration(registrationLink string) ModuleRegistration {
	return ModuleRegistration{
		registrationLink: registrationLink,
		session:          session,
	}
}
