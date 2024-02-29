package moduleGetter

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"strings"
)

type Category struct {
	Title      string     // Title of the Category e.g. "Compulsory Modules Informatics"
	Url        string     // Link associated to title anchor
	Categories []Category // All categories, which are listed under the current category
	Modules    []Module   // All Module's the category contains
}

// Module represents a module open for registration.
type Module struct {
	Title            string // Title of the module
	Teacher          string // Teachers of the module
	RegistrationLink string // The link a  user gets re-directed to, if he clicks the red "Register" button
	Events           []Event
}

// Event represents events of a module like exercises or lectures.
type Event struct {
	Id              string // ID of the event in the following format 64-010
	Title           string // Title of the event
	Link            string // The link a user gets re-directed to, if he clicks the title
	MaxCapacity     int    // Maximum student capacity of the event
	CurrentCapacity int    // Currently registered students for the event
}

/*
Refresh re-fetches the data of a module from the STiNE servers. It is only recommended when fast speed is required
or bandwidth needs to be saved. Otherwise, run GetAvailableModules again.
*/
func (module *Module) Refresh() {
	// session no should not have changed, otherwise session number in url is not valid anymore
}

func extractCategories(doc *goquery.Document) ([]Category, error) {
	var categories []Category

	// extract the category list anchor entries
	doc.Find("ul:not([class]) li a").Each(func(index int, category *goquery.Selection) {
		// title is always text inside anchor
		title := category.Text()
		// something unnecessary whitespace is added in the title at the start or end, remove
		title = strings.TrimSpace(title)
		// href is link to the category page
		link, exists := category.Attr("href")

		if !exists {
			fmt.Println("Some categories may be missing, as there was an anchor with a missing href")
		}

		categories = append(categories, Category{
			Title: title,
			Url:   link,
		})
	})

	return categories, nil
}

/*
GetAvailableModules returns the modules currently listed under "Studying" > "Register for modules and courses".

The depth indicates how deep different modules are nested within a category.

For instance, at a depth of 2, a structure like 'Computer Science' -> 'Elective Area' -> 'Module 1' would be displayed.
However, a further nested structure like 'Computer Science' -> 'Elective Area' -> 'Abroad' -> 'Module 2' would not be shown,
as it exceeds the specified depth limit.

The registerURL represents the URL, which re-directs to "Studying" > "Register for modules and courses".

The client is the HTTP Client the requests should be executed with.
*/
func GetAvailableModules(depth int, registerURL string, client *http.Client) ([]Category, error) {
	resp, errGet := client.Get(registerURL)
	if errGet != nil {
		return nil, errGet
	}

	doc, errDoc := goquery.NewDocumentFromReader(resp.Body)
	if errDoc != nil {
		return nil, errDoc
	}

	categories, errCat := extractCategories(doc)
	if errCat != nil {
		return nil, errCat
	}

	for range categories {

	}

	fmt.Println(categories)

	return []Category{}, nil
}
