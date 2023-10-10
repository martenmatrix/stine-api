package stineapi

import "fmt"

type general struct {
	MatriculationNumber string
	Name                string
	ForwardToUniEmail   bool
	SecondCitizenship   string
	Phone               string
	Mobile              string
	Mail                string
	UniMail             string
}

type address struct {
	Street          string
	AddressAddition string
	Country         string
	PostalCode      string
	City            string
}

type statistics struct {
	GermanState string
}

// UserData contains general information about the current authenticated user. It represents the information located under the "Benutzerkonto" tab.
type UserData struct {
	General    general
	Address    address
	Statistics statistics
}

func (session *Session) getUserAccountURL() string {
	return fmt.Sprintf("https://stine.uni-hamburg.de/scripts/mgrqispi.dll?APPNAME=CampusNet&PRGNAME=PERSADDRESS&ARGUMENTS=-N%s,-N000273,", session.sessionNo)
}

func (session *Session) GetUserData() (UserData, error) {
	return UserData{}, nil
}
