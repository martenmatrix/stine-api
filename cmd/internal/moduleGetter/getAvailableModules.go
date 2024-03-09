package moduleGetter

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/martenmatrix/stine-api/cmd/internal/stineURL"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

type Category struct {
	Title      string       // Title of the Category e.g. "Compulsory Modules Informatics"
	Url        string       // Link associated to title anchor
	Categories []Category   // All categories, which are listed under the current category
	Modules    []Module     // All Module's the category contains
	clientUsed *http.Client // The client used for the initial request
}

// Module represents a module open for registration.
type Module struct {
	Title            string  // Title of the module
	Teacher          string  // Teachers of the module
	RegistrationLink string  // Link a user gets re-directed to, if he wants to register for the module. It will return an empty string, if the user has already registered for the module
	Events           []Event // All events, which are correlated to the module like exercises and lectures
}

// Event represents events of a module like exercises or lectures.
type Event struct {
	Id              string  // ID of the event in the following format 64-010
	Title           string  // Title of the event
	Link            string  // The link a user gets re-directed to, if he clicks the title
	MaxCapacity     float64 // Maximum student capacity of the event
	CurrentCapacity float64 // Currently registered students for the event
}

// check if string already contains http for testing purposes, otherwise add stine url before path
func addSTiNEPrefix(path string) string {
	if path == "" {
		return ""
	} else if strings.Contains(path, "http://") {
		// do not add anything
		return path
	} else {
		// add stine path
		return stineURL.Url + path
	}
}

func extractCategories(doc *goquery.Document) ([]Category, error) {
	var categories []Category

	// extract the category list anchor entries
	doc.Find("ul:not([class]) li a").Each(func(index int, category *goquery.Selection) {
		// title is always text inside anchor
		title := category.Text()
		// sometimes unnecessary whitespace is added in the title at the start or end, remove
		title = strings.TrimSpace(title)
		// href is link to the category page
		link, exists := category.Attr("href")

		if !exists {
			fmt.Println("Some categories may be missing, as there was an anchor with a missing href")
		}

		categories = append(categories, Category{
			Title: title,
			Url:   addSTiNEPrefix(link),
		})
	})

	return categories, nil
}

func isEvent(eventSelection *goquery.Selection) bool {
	html, err := eventSelection.Html()
	if err != nil {
		fmt.Println("Could not parse HTML and evaluate if selection is an event, returning false")
		return false
	}

	return strings.Contains(html, "<!--logo column-->")
}

func extractEvent(eventSelection *goquery.Selection) (Event, error) {
	paragraphs := eventSelection.Find("p")

	id := paragraphs.Find("a[name='eventLink']").Text()[0:6]

	title := paragraphs.Find(".eventTitle").Text()
	// something unnecessary whitespace is added in the title at the start or end, remove
	title = strings.TrimSpace(title)

	link, exists := eventSelection.Find("a").Attr("href")
	if !exists {
		link = ""
	}

	// extract capacity with regex
	capacityString, htmlErr := eventSelection.Find(".tbdata:has(br)").Html()
	if htmlErr != nil {
		return Event{}, htmlErr
	}
	placesReg := regexp.MustCompile("(\\d+|-) \\| (\\d+|-)") // regex matches "number OR - | number OR -"
	dataWithoutDate := placesReg.FindString(capacityString)
	dataWithoutWhitespace := strings.ReplaceAll(dataWithoutDate, " ", "")
	dataInSlice := strings.Split(dataWithoutWhitespace, "|")

	maxCapString := dataInSlice[0]
	usedCapString := dataInSlice[1]

	var maxCap float64
	var usedCap float64

	if maxCapString == "-" {
		maxCap = math.Inf(1)
	} else {
		var maxCapErr error
		maxCap, maxCapErr = strconv.ParseFloat(maxCapString, 64)
		if maxCapErr != nil {
			return Event{}, maxCapErr
		}
	}

	if usedCapString == "-" {
		usedCap = math.Inf(1)
	} else {
		var usedCapErr error
		usedCap, usedCapErr = strconv.ParseFloat(usedCapString, 64)
		if usedCapErr != nil {
			return Event{}, usedCapErr
		}
	}

	return Event{
		Id:              id,
		Title:           title,
		Link:            addSTiNEPrefix(link),
		MaxCapacity:     maxCap,
		CurrentCapacity: usedCap,
	}, nil
}

func extractEvents(moduleHeading *goquery.Selection) ([]Event, error) {
	var events []Event

	// get all following trs, until next module starts, those are the events
	// initial tr is not included
	modules := moduleHeading.NextUntil("tr:has(td.tbsubhead)")
	modules.Each(func(i int, selection *goquery.Selection) {
		// do not iterate over title headings from modules
		if isEvent(selection) {
			event, err := extractEvent(selection)
			if err != nil {
				fmt.Println("Unable to parse an event, skipping")
			} else {
				events = append(events, event)
			}
		}
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
				RegistrationLink: addSTiNEPrefix(registerLink),
				Events:           events,
			})
		}
	})

	return modules, nil
}

func getCategory(client *http.Client, title string, url string) (Category, error) {
	var category Category

	// fetch new site category links to
	resp, errGet := client.Get(url)
	if errGet != nil {
		return Category{}, errGet
	}

	// convert to goquery doc
	doc, docErr := goquery.NewDocumentFromReader(resp.Body)
	if docErr != nil {
		return Category{}, docErr
	}

	// extract categories from newly fetched page and set as new categories, so while loop keeps iterating over them
	containsCategories, errCat := extractCategories(doc)
	if errCat != nil {
		return Category{}, errCat
	}

	// extract modules from newly fetched page
	containsModules, errMod := extractModules(doc)
	if errMod != nil {
		return Category{}, errMod
	}

	category.Title = title
	category.Url = url
	// set categories and modules of current category
	category.Categories = containsCategories
	category.Modules = containsModules
	category.clientUsed = client

	return category, nil
}

var depth int

// recursively gets all child categories of the passed category and returns the edited passed category struct
func getChildCategories(client *http.Client, category Category, maxDepth int) (Category, error) {
	if depth >= maxDepth {
		// break
		return category, nil
	}

	var childCategories []Category

	for _, category := range category.Categories {
		// needs to be child of prev category
		parsedCategory, parseErr := getCategory(client, category.Title, category.Url)
		if parseErr != nil {
			return Category{}, parseErr
		}

		// parse every child category and add it to struct
		childCategories = append(childCategories, parsedCategory)

		depth++
		_, err := getChildCategories(client, parsedCategory, maxDepth)
		if err != nil {
			return Category{}, err
		}
	}
	// store new categories as child
	category.Categories = childCategories

	return category, nil
}

/*
GetAvailableModules returns the modules currently listed under "Studying" > "Register for modules and courses".

The depth indicates how deep different categories are nested within a category.

For instance, at a depth of 2, a structure like 'Computer Science' -> 'Elective Area' -> 'Module 1' would be displayed.
However, a further nested structure like 'Computer Science' -> 'Elective Area' -> 'Abroad' -> 'Module 2' would not be shown,
as it exceeds the specified depth limit.

The registerURL represents the URL, which re-directs to "Studying" > "Register for modules and courses".

The client is the HTTP Client the requests should be executed with.
*/
func GetAvailableModules(depth int, registerURL string, client *http.Client) (Category, error) {
	// handle first page
	firstCategory, firstCatErr := getCategory(client, "initial", registerURL)
	if firstCatErr != nil {
		return Category{}, firstCatErr
	}

	withSubCategories, err := getChildCategories(client, firstCategory, depth)
	if err != nil {
		return Category{}, nil
	}

	return withSubCategories, nil
}

/*
Refresh re-fetches the data of a category from the STiNE servers. The session of the initial request cannot be expired.

In order to refresh a module, the whole category needs to be re-fetched, which makes a Refresh function for a module useless.

The client is the HTTP client used for the request, needs to be authenticated on STiNE.

The depth indicates how deep different categories are nested within a category.
*/
func (category *Category) Refresh(depth int) (Category, error) {
	// handle first page
	firstCategory, firstCatErr := getCategory(category.clientUsed, category.Title, category.Url)
	if firstCatErr != nil {
		return Category{}, firstCatErr
	}

	withSubCategories, err := getChildCategories(category.clientUsed, firstCategory, depth)
	if err != nil {
		return Category{}, nil
	}

	return withSubCategories, nil
}
