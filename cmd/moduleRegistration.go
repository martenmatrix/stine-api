package stineapi

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/martenmatrix/stine-api/cmd/internal/moduleRegisterer"
	"github.com/martenmatrix/stine-api/cmd/internal/onPage"
	"github.com/martenmatrix/stine-api/cmd/internal/sessionNo"
	"log"
	"net/http"
	"strings"
)

/*
ModuleRegistration represents a running registration for a module on the STiNE platform.
*/
type ModuleRegistration struct {
	registrationLink string
	registrationId   string // id from a hidden input field, which is returned after requesting the registrationLink
	menuId           string // menu id represents, which option is selected on the menu to the left on the stine page
	ExamDate         int    // The selected exam date, 0 - first exam, 1 - second exam, or 2 - another semester
	sessionNumber    string
	client           *http.Client
}

// creates a TanRequired struct for the required iTAN
func (modReg *ModuleRegistration) getTanRequiredStruct(doc *goquery.Document) *TanRequired {
	itanStart := doc.Find(".itan").First().Text()

	iTANWithLeadingZero := strings.ReplaceAll(itanStart, " ", "0")

	return &TanRequired{
		client:         modReg.client,
		sessionNo:      modReg.sessionNumber,
		url:            modReg.registrationLink,
		registrationId: modReg.registrationId,
		TanStartsWith:  iTANWithLeadingZero,
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
		log.Println(fmt.Sprintf("SetExamDate only accepts the integers from 1 to 2. The exam date (current value: %d) will not be changed.", examDate))
	} else {
		modReg.ExamDate = examDate
	}
}

/*
Register sends the registration to the STiNE servers.
If an iTAN is required, instead of nil a [TanRequired] is returned.
*/
func (modReg *ModuleRegistration) Register(client *http.Client, sessionNumber string) (*TanRequired, error) {
	modReg.client = client
	modReg.sessionNumber = sessionNumber

	var currentResponse *http.Response
	var currentDocument *goquery.Document
	var err error

	modReg.registrationLink = sessionNo.Refresh(modReg.registrationLink, sessionNumber)
	regId, err := moduleRegisterer.GetRegistrationId(client, modReg.registrationLink)
	if err != nil {
		return nil, err
	}

	currentResponse, err = moduleRegisterer.DoRegistrationRequest(client, modReg.registrationLink, sessionNumber, modReg.menuId, regId)
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
	if onPage.OnSelectExamPage(currentDocument) {
		rbCode, rbErr := moduleRegisterer.GetRBCode(currentDocument)
		if rbErr != nil {
			return nil, rbErr
		}
		currentResponse, err = moduleRegisterer.DoExamRegistrationRequest(client, modReg.registrationLink, rbCode, sessionNumber, modReg.menuId, regId, modReg.ExamDate)
		if err != nil {
			return nil, err
		}
		currentDocument, err = goquery.NewDocumentFromReader(currentResponse.Body)
		if err != nil {
			return nil, err
		}
	}

	if onPage.OniTANPage(currentDocument) {
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
func CreateModuleRegistration(registrationLink string) *ModuleRegistration {
	return &ModuleRegistration{
		registrationLink: registrationLink,
		menuId:           "000309",
	}
}
