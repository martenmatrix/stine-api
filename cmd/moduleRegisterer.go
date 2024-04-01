package stineapi

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/martenmatrix/stine-api/cmd/internal/onPage"
	"github.com/martenmatrix/stine-api/cmd/internal/sessionNo"
	"log"
	"net/http"
	"net/url"
	"strings"
)

// DoRegistrationRequest initiates the registration request on the STiNE servers
func doRegistrationRequest(client *http.Client, reqUrl string, sessionNo string, menuId string, registrationId string) (*http.Response, error) {
	formQuery := url.Values{
		"Next":      {" Weiter"},
		"APPNAME":   {"CampusNet"},
		"PRGNAME":   {"SAVEREGISTRATIONDETAILS"},
		"ARGUMENTS": {"sessionno,menuid,rgtr_id"},
		"sessionno": {sessionNo},
		"menuid":    {menuId},
		"rgtr_id":   {registrationId},
	}
	res, err := client.PostForm(reqUrl, formQuery)

	if err != nil {
		return nil, err
	}
	return res, nil
}

// GetRBCode extracts the RB-Code from sites, where user needs to select an exam date
func getRBCode(doc *goquery.Document) (string, error) {
	// on all pages where a user is able to select an exam date, every input has a name attribute with the same id (called rb code because the id starts with RB_)
	rbCode, exists := doc.Find(`input[type="radio"]`).First().Attr("name")
	if !exists {
		return "", errors.New("name attribute with rb code does not exist on input")
	}
	return rbCode, nil
}

// converts the selected exam to the stine exam date types
func getExamMode(examDate int) string {
	switch examDate {
	case 0:
		return " 1"
	case 1:
		return " 2"
	case 2:
		return "99"
	}
	return " 1"
}

// DoExamRegistrationRequest sends the exam selection to the servers, this only works after DoRegistrationRequest was executed
func doExamRegistrationRequest(client *http.Client, reqUrl string, rbCode string, sessionNo string, menuId string, registrationId string, examDate int) (*http.Response, error) {
	formQuery := url.Values{
		"Next":      {" Next"},
		rbCode:      {getExamMode(examDate)},
		"APPNAME":   {"CAMPUSNET"},
		"PRGNAME":   {"SAVEEXAMDETAILS"},
		"ARGUMENTS": {"sessionno,menuid,rgtr_id,mode"},
		"sessionno": {sessionNo},
		"menuid":    {menuId},
		"rgtr_id":   {registrationId},
		"mode":      {"0001"},
	}

	res, err := client.PostForm(reqUrl, formQuery)

	if err != nil {
		return nil, err
	}

	return res, nil
}

// GetRegistrationId extracts the registrationId from the HTML, which the registrationLink links to
func getRegistrationId(client *http.Client, registrationLink string) (string, error) {
	res, _ := client.Get(registrationLink)
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
func (modReg *ModuleRegistration) Register() (*TanRequired, error) {

	var currentResponse *http.Response
	var currentDocument *goquery.Document
	var err error

	modReg.registrationLink = sessionNo.Refresh(modReg.registrationLink, modReg.sessionNumber)
	regId, err := getRegistrationId(modReg.client, modReg.registrationLink)
	if err != nil {
		return nil, err
	}

	currentResponse, err = doRegistrationRequest(modReg.client, modReg.registrationLink, modReg.sessionNumber, modReg.menuId, regId)
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
		rbCode, rbErr := getRBCode(currentDocument)
		if rbErr != nil {
			return nil, rbErr
		}
		currentResponse, err = doExamRegistrationRequest(modReg.client, modReg.registrationLink, rbCode, modReg.sessionNumber, modReg.menuId, regId, modReg.ExamDate)
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

The registrationLink is the url the red "Register" button links to on the STiNE page.
*/
func createModuleRegistration(registrationLink string, sessionNumber string, client *http.Client) *ModuleRegistration {
	return &ModuleRegistration{
		registrationLink: registrationLink,
		sessionNumber:    sessionNumber,
		client:           client,
		menuId:           "000309",
	}
}
