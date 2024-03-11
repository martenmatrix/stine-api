package stineapi

import (
	"fmt"
	"github.com/martenmatrix/stine-api/cmd/internal/moduleGetter"
)

func ExampleSession_Login() {
	// Authenticate user
	session := NewSession()
	err := session.Login("BBB????", "password")

	if err != nil {
		fmt.Println("Authentication failed")
	}

	// Session is now authenticated
	fmt.Println(session.SessionNo) // returns e.g. 631332205304636
}

func ExampleSession_GetCategories() {
	// Session should be authenticated
	session := NewSession()

	initialCategory, err := session.GetCategories(1)

	// Print title of every category reachable from first category
	for _, category := range initialCategory.Categories {
		fmt.Println(category.Title)
	}

	vssModule := initialCategory.Categories[0].Modules[1] // select "Distributed Systems and Systems Security (SuSe 23)" module located at second place in first listed category

	fmt.Println(vssModule.Title)   // Distributed Systems and Systems Security (SuSe 23)
	fmt.Println(vssModule.Teacher) // Prof. Dr. Name Surname

	firstEvent := vssModule.Events[0]

	fmt.Println(fmt.Printf("Available places: %f", firstEvent.MaxCapacity))   // print places available
	fmt.Println(fmt.Printf("Booked places : %f", firstEvent.CurrentCapacity)) // print places already booked

	// Refresh everything listed under initialCategory.Categories[0]
	// Only works on categories, all modules within a category will be refreshed
	firstCategoryRefresh, err := initialCategory.Categories[0].Refresh(0)
	if err != nil {
		// Handle error
	}

	// Check e.g., if places became available
	fmt.Println(firstCategoryRefresh)
}

func ExampleSession_RegisterForModule() {
	// Session should be authenticated
	session := NewSession()

	// Module ideally should be retrieved with GetCategories
	vssModule := moduleGetter.Module{}

	// Create module registration
	moduleRegistration := session.RegisterForModule(vssModule)
	moduleRegistration.SetExamDate(1)            // Select second available exam date
	tanReq, err := moduleRegistration.Register() // Send registration to servers

	if err != nil {
		// Handle error
	}

	if tanReq != nil {
		// iTAN is required for registration
		fmt.Println(tanReq.TanStartsWith) // Print starting numbers of itan e.g. 087
		err := tanReq.SetTan("087233233") // We can also enter the tan without prefix e.g. 233233

		if err != nil {
			// Handle error
		}
	}
	// User is registered for the module and maybe also registered for the exam, sometimes you are only able to select an exam after joining the lecture
}
