package moduleRegisterer

import (
	"errors"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"net/url"
)

// DoRegistrationRequest initiates the registration request on the STiNE servers
func DoRegistrationRequest(client *http.Client, reqUrl string, sessionNo string, menuId string, registrationId string) (*http.Response, error) {
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
func GetRBCode(doc *goquery.Document) (string, error) {
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
func DoExamRegistrationRequest(client *http.Client, reqUrl string, rbCode string, sessionNo string, menuId string, registrationId string, examDate int) (*http.Response, error) {
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
func GetRegistrationId(client *http.Client, registrationLink string) (string, error) {
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
