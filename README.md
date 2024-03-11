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
- :white_check_mark: User Auth
- :white_check_mark: Fetch categories available for user
- :white_check_mark: Fetch modules available for user
- :white_check_mark: Register user for a module


- :negative_squared_cross_mark: Fetch schedules for a user
- :negative_squared_cross_mark: Register user for a lecture
- :negative_squared_cross_mark: Register user for an exercise
- :negative_squared_cross_mark: Get information about the user
- :negative_squared_cross_mark: Get exam results for the user
- :negative_squared_cross_mark: Get exam results for the user by using the mobile STiNE API Endpoint (results show up earlier)

## :paperclip: Examples
### Authenticate a user
```go
package main

import "github.com/martenmatrix/stine-api/cmd"
import "fmt"

func main() {
	// Authenticate user
	session := stineapi.NewSession()
	err := session.Login("BBB????", "password")

	if err != nil {
		fmt.Println("Authentication failed")
	}

	// session is now authenticated
	fmt.Println(session.SessionNo) // returns e.g. 631332205304636
}
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