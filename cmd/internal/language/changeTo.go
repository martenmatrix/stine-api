package language

import (
	"github.com/martenmatrix/stine-api/cmd/internal/sessionNo"
	"github.com/martenmatrix/stine-api/cmd/internal/stineURL"
	"net/http"
)

const (
	englishLink = stineURL.Url + "/scripts/mgrqispi.dll?APPNAME=CampusNet&PRGNAME=CHANGELANGUAGE&ARGUMENTS=-N000000000000000,-N002"
	germanLink  = stineURL.Url + "/scripts/mgrqispi.dll?APPNAME=CampusNet&PRGNAME=CHANGELANGUAGE&ARGUMENTS=-N000000000000000,-N001"
)

/*
ChangeToEnglish changes the language to english on the STiNE website.
*/
func ChangeToEnglish(client *http.Client, sessionNumber string) error {
	_, err := client.Get(sessionNo.Refresh(englishLink, sessionNumber))

	if err != nil {
		return err
	}

	return nil
}

/*
ChangeToGerman changes the language to german on the STiNE website.
*/
func ChangeToGerman(client *http.Client, sessionNumber string) error {
	_, err := client.Get(sessionNo.Refresh(germanLink, sessionNumber))

	if err != nil {
		return err
	}

	return nil
}
