package moduleGetter

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"strings"
)

type Category struct {
	Title      string      // Title of the Category e.g. "Compulsory Modules Informatics"
	Url        string      // Link associated to title anchor
	Categories *[]Category // All categories, which are listed under the current category
	Modules    *[]Module   // All Module's the category contains
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

func extractEvents(moduleHeading *goquery.Selection) ([]Event, error) {
	var events []Event

	// get all following trs, until next module starts, those are the events
	// initial tr is not included
	modules := moduleHeading.NextUntil("tr:has(td.tbsubhead)")
	modules.Each(func(i int, selection *goquery.Selection) {
	})

	return events, nil
}

func extractModules(doc *goquery.Document) ([]Module, error) {
	var modules []Module

	doc.Find("tr").Each(func(i int, selection *goquery.Selection) {
		html, err := selection.Html()
		if err != nil {
			fmt.Println("Some modules may be missing, as HTML could not be parsed")
		}

		// only select those trs, which are the heading of a module (they contain <!-- MODULE --> as a html comment)
		if strings.Contains(html, "<!-- MODULE -->") {
			// iterate over each module
			title := selection.Find(".eventTitle").Text()
			teacher := selection.Find("p:not(:has(a))").Text()
			registerLink, exists := selection.Find(".register").Attr("href")
			if !exists {
				registerLink = ""
			}
			events, err := extractEvents(selection)
			if err != nil {
				fmt.Println(fmt.Sprintf("The events associated to the module %s could not be extracted", title))
				events = []Event{}
			}

			modules = append(modules, Module{
				Title:            title,
				Teacher:          teacher,
				RegistrationLink: registerLink,
				Events:           events,
			})
		}
	})

	return modules, nil
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
func GetAvailableModules(depth int, registerURL string, client *http.Client) (Category, error) {
	initialCategory := Category{
		Title: "initialPage",
		Url:   registerURL,
	}

	resp, errGet := client.Get(registerURL)
	if errGet != nil {
		return Category{}, errGet
	}

	doc, errDoc := goquery.NewDocumentFromReader(resp.Body)
	if errDoc != nil {
		return Category{}, errDoc
	}

	// handle first page
	categories, errCat := extractCategories(doc)
	if errCat != nil {
		return Category{}, errCat
	}

	modules, errMod := extractModules(doc)
	if errMod != nil {
		return Category{}, errMod
	}

	// save first categories and modules
	initialCategory.Categories = &categories
	initialCategory.Modules = &modules

	currDepth := 0

	// while there are categories left, traverse trough them
	for len(categories) > 0 && currDepth < depth {
		currDepth++
		// iterate over every category in category list
		for _, category := range categories {
			// store old category
			oldCategory := category

			// fetch new site category links to
			resp, errGet := client.Get(category.Url)
			if errGet != nil {
				return Category{}, errGet
			}

			// convert to goquery doc
			newDoc, newDocErr := goquery.NewDocumentFromReader(resp.Body)
			if newDocErr != nil {
				return Category{}, newDocErr
			}

			// extract categories from newly fetched page and set as new categories, so while loop keeps iterating over them
			categories, errCat = extractCategories(newDoc)
			if errCat != nil {
				return Category{}, errCat
			}

			// extract modules from newly fetched page
			modules, errMod := extractModules(newDoc)
			if errMod != nil {
				return Category{}, errMod
			}

			// set categories and modules of current category
			oldCategory.Categories = &categories
			oldCategory.Modules = &modules
		}
	}

	return initialCategory, nil
}
