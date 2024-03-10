package onPage

import (
	"github.com/PuerkitoBio/goquery"
	"log"
)

// OniTANPage checks, if the HTML of the response asks for an iTAN
func OniTANPage(doc *goquery.Document) bool {
	return doc.Find(".itan").Length() > 0
}

// OnSelectExamPage checks, if the HTML of the response asks the user to select an exam
func OnSelectExamPage(doc *goquery.Document) bool {
	inputValue, exists := doc.Find(`input[name="PRGNAME"]`).Attr("value")
	if !exists {
		log.Println("Could not evaluate, if an exam selection is required. Returning false.")
		return false
	}
	return inputValue == "SAVEEXAMDETAILS"
}
