package stineapi

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"net/url"
	"strings"
)

/*
ModuleRegistration represents a running registration for a module on the STiNE platform.
*/
type ModuleRegistration struct {
	registrationLink string
	registrationId   string // id from a hidden input field, which is returned after requesting the registrationLink
	menuId           string // menu id represents, which option is selected on the menu to the left on the stine page
	examDate         int
	session          *Session
}

/*
TanRequired is returned from a function, if an iTAN is needed to complete the action.
TanStartsWith represents the two starting numbers of the required iTAN with a leading zero.
*/
type TanRequired struct {
	registration  *ModuleRegistration
	TanStartsWith string
}

func checkForTANError(res *http.Response) error {
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return err
	}

	errorMsg := doc.Find(".error").First().Text()
	if errorMsg != "" {
		return errors.New(fmt.Sprintf("itan validation could not be completed: %s", errorMsg))
	}
	return nil
}

func (tanReq *TanRequired) sendTAN(reqURL string, itanWithoutPrefix string) error {
	formQuery := url.Values{
		"campusnet_submit": {""},
		"tan_code":         {itanWithoutPrefix},
		"APPNAME":          {"CampusNet"},
		"PRGNAME":          {"SAVEREGISTRATION"},
		"ARGUMENTS":        {"sessionno,menuid,rgtr_id,mode,timetable_id,location_id"},
		"sessionno":        {tanReq.registration.session.sessionNo},
		"rgtr_id":          {tanReq.registration.registrationId},
		"mode":             {"   0"},
	}
	res, err := tanReq.registration.session.Client.PostForm(reqURL, formQuery)
	defer res.Body.Close()

	if err != nil {
		return err
	}

	tanErr := checkForTANError(res)
	if tanErr != nil {
		return tanErr
	}

	return nil
}

func (tanReq *TanRequired) removeTanPrefix(itan string) string {
	tanWithoutPrefix, _ := strings.CutPrefix(itan, tanReq.TanStartsWith)
	return tanWithoutPrefix
}

/*
SetTan sends the provided iTAN to the STiNE servers to complete an action. If the validation fails, an error is returned.
The users iTAN list will be disabled after 3 failed attempts.
The iTAN can be entered with the [TanStartsWith] as a prefix or without.
*/
func (tanReq *TanRequired) SetTan(itan string) error {
	return nil
}

func (modReg *ModuleRegistration) doRegistrationRequest(reqUrl string) (*http.Response, error) {
	formQuery := url.Values{
		"Next":      {" Weiter"},
		"APPNAME":   {"CampusNet"},
		"PRGNAME":   {"SAVEREGISTRATIONDETAILS"},
		"ARGUMENTS": {"sessionno,menuid,rgtr_id"},
		"sessionno": {modReg.session.sessionNo},
		"menuid":    {modReg.menuId},
		"rgtr_id":   {modReg.registrationId},
	}
	res, err := modReg.session.Client.PostForm(reqUrl, formQuery)

	if err != nil {
		return nil, err
	}
	return res, nil
}

// on all pages where a user is able to select an exam date, every input has a name attribute with the same id (called rb code because the id starts with RB_)
func getRBCode(doc *goquery.Document) (string, error) {
	rbCode, exists := doc.Find(`input[type="radio"]`).First().Attr("name")
	if !exists {
		return "", errors.New("name attribute with rb code does not exist on input")
	}
	return rbCode, nil
}

func (modReg *ModuleRegistration) getExamMode() string {
	switch modReg.examDate {
	case 0:
		return " 1"
	case 1:
		return " 2"
	case 2:
		return "99"
	}
	return " 1"
}

func (modReg *ModuleRegistration) doExamRegistrationRequest(reqUrl string, rbCode string) (*http.Response, error) {
	formQuery := url.Values{
		"Next":      {" Next"},
		rbCode:      {modReg.getExamMode()},
		"APPNAME":   {"CAMPUSNET"},
		"PRGNAME":   {"SAVEEXAMDETAILS"},
		"ARGUMENTS": {"sessionno,menuid,rgtr_id,mode"},
		"sessionno": {modReg.session.sessionNo},
		"menuid":    {modReg.menuId},
		"rgtr_id":   {modReg.registrationId},
		"mode":      {"0001"},
	}

	res, err := modReg.session.Client.PostForm(reqUrl, formQuery)

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (modReg *ModuleRegistration) getRegistrationId() error {
	res, _ := modReg.session.Client.Get(modReg.registrationLink)
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

func oniTANPage(doc *goquery.Document) bool {
	return doc.Find(".itan").Length() > 0
}

func onSelectExamPage(doc *goquery.Document) bool {
	inputValue, exists := doc.Find(`input[name="PRGNAME"]`).Attr("value")
	if !exists {
		log.Println("Could not evaluate, if an exam selection is required. Returning false.")
		return false
	}
	return inputValue == "SAVEEXAMDETAILS"
}

func (modReg *ModuleRegistration) getTanRequiredStruct(doc *goquery.Document) *TanRequired {
	itanStart := doc.Find(".itan").First().Text()

	iTANWithLeadingZero := strings.ReplaceAll(itanStart, " ", "0")

	return &TanRequired{
		registration:  modReg,
		TanStartsWith: iTANWithLeadingZero,
	}
}

/*
SetExamDate allows you to choose a specific exam date for the initial registration. If this function is not executed, the first exam date is selected by default.

The exam date will not be changed, if the user is already registered for the module.

0 - Selects the first exam date (default choice).

1 - Selects the second exam date.

2 - Opts for writing the exam in a different semester (exact date not specified).
*/
func (modReg *ModuleRegistration) SetExamDate(examDate int) {
	if examDate < 0 || examDate > 2 {
		log.Println(fmt.Sprintf("SetExamDate only accepts the integers from 1 to 2. The exam date (current value: %d) will not be changed.", modReg.examDate))
	} else {
		modReg.examDate = examDate
	}
}

/*
Register sends the registration to the STiNE servers.
If an iTAN is required, instead of nil a [TanRequired] is returned.
*/
func (modReg *ModuleRegistration) Register() (*TanRequired, error) {
	var currentResponse *http.Response
	var currentDocument *goquery.Document
	var err error

	modReg.registrationLink = modReg.session.RefreshSessionNumber(modReg.registrationLink)
	err = modReg.getRegistrationId()
	if err != nil {
		return nil, err
	}

	currentResponse, err = modReg.doRegistrationRequest(modReg.registrationLink)
	if err != nil {
		return nil, err
	}
	defer currentResponse.Body.Close()

	currentDocument, err = goquery.NewDocumentFromReader(currentResponse.Body)
	if err != nil {
		return nil, err
	}

	// only some modules require an exam registration, before the module registration can be completed
	// for some modules the exam needs to be booked, after registering for the module
	if onSelectExamPage(currentDocument) {
		rbCode, rbErr := getRBCode(currentDocument)
		if rbErr != nil {
			return nil, rbErr
		}
		currentResponse, err = modReg.doExamRegistrationRequest(modReg.registrationLink, rbCode)
		if err != nil {
			return nil, err
		}
		currentDocument, err = goquery.NewDocumentFromReader(currentResponse.Body)
		if err != nil {
			return nil, err
		}
	}

	if oniTANPage(currentDocument) {
		tan := modReg.getTanRequiredStruct(currentDocument)
		return tan, nil
	}

	return nil, nil
}

/*
CreateModuleRegistration creates and returns a [ModuleRegistration], which provides functions to register for the specified module and its corresponding events.

This function requires a registration link as an argument, which can be retrieved the following way for a specific module from the STiNE website:

1. Navigate to STiNE and login.

2. Navigate to the module subsection, where your module is listed (e.g. for Software Development I when studying Computer Science, go to "Studying" > "Register for modules and courses" > "Compulsory Modules Informatics")

3. Your module should now be displayed with a bunch of other modules.

4. There should be a red "Register" button to the right of the module name.

5. Right-click the button and click "Copy link address", this is the registration link for the module!

If there is no "Register" button, you've either already completed the module or you've already signed up for the module.
*/
func (session *Session) CreateModuleRegistration(registrationLink string) *ModuleRegistration {
	return &ModuleRegistration{
		registrationLink: registrationLink,
		session:          session,
		menuId:           "000309",
	}
}
