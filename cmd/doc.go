/*
This is an unofficial STiNE API for Go. It is easy to use, completely request-based and uses no browser automation, which makes it incredible fast.

Basic Usage:

	session := stineapi.NewSession()

	err := session.Login("user", "pass")
	if err != nil {
	// handle error
	}

	// our client is authenticated on the stine website, session provides various functions to interact with it

*/

package cmd
