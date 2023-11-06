package stineapi

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"net/url"
	"strings"
)

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

/*
TanRequired is returned from a function, if an iTAN is needed to complete the action.
TanStartsWith represents the two starting numbers of the required iTAN with a leading zero.
*/
type TanRequired struct {
	registration  *ModuleRegistration
	TanStartsWith string
}

func (tanReq *TanRequired) removeTanPrefix(itan string) string {
	tanWithoutPrefix, _ := strings.CutPrefix(itan, tanReq.TanStartsWith)
	return tanWithoutPrefix
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
The iTAN can be entered with the first three numbers or without the prefix provided by STiNE.
*/
func (tanReq *TanRequired) SetTan(itan string) error {
	tanWithoutPrefix := tanReq.removeTanPrefix(itan)
	err := tanReq.sendTAN(tanReq.registration.registrationLink, tanWithoutPrefix)
	if err != nil {
		return err
	}
	return nil
}
