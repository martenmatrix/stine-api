package stineapi

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/url"
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
	examDate         int
	session          *Session
	TanRequired      bool
	TanStartsWith    string
}

func (modReg *ModuleRegistration) refreshSessionNumber() {
	reg := regexp.MustCompile("ARGUMENTS=-N\\d{15}")
	linkWithRefreshedSessionNo := reg.ReplaceAllString(modReg.registrationLink, "ARGUMENTS=-N"+modReg.session.sessionNo)
	modReg.registrationLink = linkWithRefreshedSessionNo
}

func (session *Session) doModuleRegistrationFormRequest(reqUrl string, menuId string, registrationId string) string {
	formQuery := url.Values{
		"Next":      {"Weiter"},
		"APPNAME":   {"CAMPUSNET"},
		"PRGNAME":   {"SAVEREGISTRATIONDETAILS"},
		"ARGUMENTS": {"sessionno,menuid,rgtr_id"},
		"sessionno": {session.sessionNo},
		"menuid":    {menuId},
		"rgtr_id":   {registrationId},
	}
	res, _ := session.client.PostForm(reqUrl, formQuery)
	logResponse(res)
	return ""
}

func (modReg *ModuleRegistration) getRegistrationId() error {
	res, _ := modReg.session.client.Get(modReg.registrationLink)
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return err
	}

	regId, onPage := doc.Find(`input[name="rgtr_id"]`).First().Attr("value")
	if !onPage {
		return errors.New("unable to find registration id in response")
	}

	modReg.registrationId = regId

	return nil
}

/*
SetExamDate allows you to choose a specific exam date. If this function is not executed, the first exam date is selected by default.
0 - Selects the first exam date (default choice).
1 - Selects the second exam date.
2 - Opts for writing the exam in a different semester (exact date not specified).
*/
func (modReg *ModuleRegistration) SetExamDate(examDate int) {
	if examDate < 0 || examDate > 2 {
		log.Println(fmt.Sprintf("SetExamDate only accepts the integers from 1 to 2. The exam date (current value: %d) will not be changed.", modReg.examDate))
	}
	modReg.examDate = examDate
}

/*
CreateModuleRegistration creates and returns a ModuleRegistration, which provides functions to register for the specified module and its corresponding events.

This function requires a registration link as an argument, which can be retrieved the following way for a specific module from the STiNE website:
1. Navigate to STiNE and login.
2. Navigate to the module subsection, where your module is listed (e.g. for Software Development I when studying Computer Science, go to "Studying" > "Register for modules and courses" > "Compulsory Modules Informatics")
3. Your module should now be displayed with a bunch of other modules.
4. There should be a red "Register" button to the right of the module name.
5. Right-click the button and click "Copy link address", this is the registration link for the module!
If there is no "Register" button, you've either already completed the module or you've already signed up for the module.
*/
func (session *Session) CreateModuleRegistration(registrationLink string) ModuleRegistration {
	return ModuleRegistration{
		registrationLink: registrationLink,
		session:          session,
	}
}
