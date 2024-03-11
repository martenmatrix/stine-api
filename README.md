<img alt="STiNE Logo" height="100" src="./stine_logo.png"/>

# A STiNE API Wrapper

This is an unofficial STiNE API for Go. It is easy to use, completely request-based and uses no browser automation, which makes it incredible fast.

## Table of Contents
- [Installation](#rocket-installation)
- [Features](#sparkles-features)
- [Examples](#paperclip-examples)
- [Documentation](#books-documentation)
- [License](#scroll-license)

## :sparkles: Features
### Done
- :white_check_mark: User Auth
- :white_check_mark: Fetch categories available for user
- :white_check_mark: Fetch modules available for user
- :white_check_mark: Register user for a module
### TODOS
- :negative_squared_cross_mark: Fetch schedules for a user
- :negative_squared_cross_mark: Register user for a lecture
- :negative_squared_cross_mark: Register user for an exercise
- :negative_squared_cross_mark: Get information about the user
- :negative_squared_cross_mark: Get messages
- :negative_squared_cross_mark: Download documents
- :negative_squared_cross_mark: Start applications
- :negative_squared_cross_mark: Use contact form
- :negative_squared_cross_mark: Get exam results for the user
- :negative_squared_cross_mark: Get exam results for the user by using the mobile STiNE API Endpoint (results show up earlier)

## :paperclip: Examples
### Authenticate a user
```go
// Authenticate user
session := stineapi.NewSession()
err := session.Login("BBB????", "password")

if err != nil {
    fmt.Println("Authentication failed")
}

// session is now authenticated
fmt.Println(session.SessionNo) // returns e.g. 631332205304636
```

### Fetch categories and modules available for user
```go
// session should be authenticated
initialCategory, err := session.GetCategories(1)

// print title of every category reachable from first category
for _, category := range initialCategory.Categories {
    fmt.Println(category.Title)
}

vssModule := initialCategory.Categories[0].Modules[1] // select "Distributed Systems and Systems Security (SuSe 23)" module located at second place in first listed category

fmt.Println(vssModule.Title) // Distributed Systems and Systems Security (SuSe 23)
fmt.Println(vssModule.Teacher) // Prof. Dr. Name Surname

fmt.Println(fmt.Printf("Available places: %f", vssModule.MaxCapacity))   // print places available
fmt.Println(fmt.Printf("Booked places : %f", vssModule.CurrentCapacity)) // print places already booked

// Refresh everything listed under initialCategory.Categories[0]
// Only works on categories, all modules within a category will be refreshed
firstCategoryRefresh, err := initialCategory.Categories[0].Refresh(0)
if err != nil {
    // Handle error
}

// Check e.g., if places became available
fmt.Println(firstCategoryRefresh)
```

## :rocket: Installation
Execute the following line in your Go project:
```shell
GOPROXY=direct go get github.com/martenmatrix/stine-api/cmd
```
and import it with
```go
import "github.com/martenmatrix/stine-api/cmd"
```

## :books: Documentation
The documentation can be found [here](https://pkg.go.dev/github.com/martenmatrix/stine-api/cmd).

## :scroll: License
[MIT](./LICENSE)