package tan

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"net/url"
	"strings"
)

// CheckForTANError checks if there was an error after entering the iTAN by reading the HTML of the response
func CheckForTANError(res *http.Response) error {
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

// RemoveTanPrefix removes the prefix, if the passed iTAN starts with it
func RemoveTanPrefix(itan string, prefix string) string {
	tanWithoutPrefix, _ := strings.CutPrefix(itan, prefix)
	return tanWithoutPrefix
}

// SendTAN does the request, which sends the iTAN to the STiNE servers, it returns an error, if the authentication was not successful
func SendTAN(client *http.Client, reqURL string, itanWithoutPrefix string, sessionNumber string, registrationId string) error {
	formQuery := url.Values{
		"campusnet_submit": {""},
		"tan_code":         {itanWithoutPrefix},
		"APPNAME":          {"CampusNet"},
		"PRGNAME":          {"SAVEREGISTRATION"},
		"ARGUMENTS":        {"sessionno,menuid,rgtr_id,mode,timetable_id,location_id"},
		"sessionno":        {sessionNumber},
		"rgtr_id":          {registrationId},
		"mode":             {"   0"},
	}
	res, err := client.PostForm(reqURL, formQuery)
	defer res.Body.Close()

	if err != nil {
		return err
	}

	tanErr := CheckForTANError(res)
	if tanErr != nil {
		return tanErr
	}

	return nil
}
