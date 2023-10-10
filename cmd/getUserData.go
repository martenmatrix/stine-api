package stineapi

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
