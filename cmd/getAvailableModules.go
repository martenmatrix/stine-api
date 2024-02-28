package stineapi

type Category struct {
	title      string      // Title of the Category e.g. "Compulsory Modules Informatics"
	url        string      // Link associated to title anchor
	categories *[]Category // All categories, which are listed under the current category
	modules    *[]Module   // All Module's the category contains
}

// Module represents a module open for registration.
type Module struct {
	title            string // Title of the module
	teacher          string // Teachers of the module
	registrationLink string // The link a  user gets re-directed to, if he clicks the red "Register" button
	events           *[]Event
}

// Event represents events of a module like exercises or lectures.
type Event struct {
	id              string // ID of the event in the following format 64-010
	title           string // Title of the event
	link            string // The link a user gets re-directed to, if he clicks the title
	maxCapacity     int    // Maximum student capacity of the event
	currentCapacity int    // Currently registered students for the event
}

/*
Refresh re-fetches the data of a module from the STiNE servers. It is only recommended when fast speed is required
or bandwidth needs to be saved. Otherwise, run GetAvailableModules again.
*/
func (module *Module) Refresh() {
	// session no should not have changed, otherwise session number in url is not valid anymore
}

/*
GetAvailableModules returns the modules currently listed under "Studying" > "Register for modules and courses".

The depth indicates how deep different modules are nested within a category.

For instance, at a depth of 2, a structure like 'Computer Science' -> 'Elective Area' -> 'Module 1' would be displayed.
However, a further nested structure like 'Computer Science' -> 'Elective Area' -> 'Abroad' -> 'Module 2' would not be shown,
as it exceeds the specified depth limit.

The registerURL represents the URL, which re-directs to "Studying" > "Register for modules and courses".
*/
func GetAvailableModules(depth int, registerURL string) {

}