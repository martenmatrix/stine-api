package stineapi

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
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
		"Next":      {"Weiter"},
		"APPNAME":   {"CAMPUSNET"},
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

func oniTANPage(res *http.Response) bool {
	htmlText, err := io.ReadAll(res.Body)
	if err != nil {
		log.Println("Could not evaluate, if an iTAN is required. Returning false.")
	}
	return strings.Contains(string(htmlText), "<!-- CONFIRM AND TAN INPUT -->")
}

func onSelectExamPage(res *http.Response) bool {
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Println("Could not evaluate, if an exam selection is required. Returning false.")
	}

	return doc.Find(`input[name="PRGNAME"]`).Text() == "SAVEEXAMDETAILS"
}

func (modReg *ModuleRegistration) getTanRequiredStruct(response *http.Response) (*TanRequired, error) {
	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return nil, err
	}

	itanStart := doc.Find(".itan").First().Text()

	iTANWithLeadingZero := strings.ReplaceAll(itanStart, " ", "0")

	return &TanRequired{
		registration:  modReg,
		TanStartsWith: iTANWithLeadingZero,
	}, nil
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
	}
	modReg.examDate = examDate
}

/*
Register sends the registration to the STiNE servers.
If an iTAN is required, instead of nil a [TanRequired] is returned.
*/
func (modReg *ModuleRegistration) Register() (*TanRequired, error) {
	modReg.registrationLink = modReg.session.RefreshSessionNumber(modReg.registrationLink)
	idErr := modReg.getRegistrationId()
	if idErr != nil {
		return nil, idErr
	}

	regRes, regErr := modReg.doRegistrationRequest(modReg.registrationLink)
	if regErr != nil {
		return nil, regErr
	}

	// TODO implement exam handling

	if oniTANPage(regRes) {
		tan, tanErr := modReg.getTanRequiredStruct(regRes)
		if tanErr != nil {
			return nil, tanErr
		}
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
