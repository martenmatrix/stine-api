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
