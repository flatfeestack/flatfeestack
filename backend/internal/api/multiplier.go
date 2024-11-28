package api

import (
	"backend/internal/db"
	"fmt"
	"net/http"
)

func GetMultiplier(w http.ResponseWriter, r *http.Request, _ *db.UserDetail) {
	fmt.Println("Working")
}

// Goal of successfully getting the multiplier
/*
Prerequisit:
1. Repos get donated to
2. Repos are supported by foundations
3. Repos are tagged as trusted by flatfeestack

Algorithm steps:
Every calculation of the multiplier value is dependent on
	a) if multiple repositories are selected for donation
	b) if at least one of these repositories are trusted by flatfeestack
1. User selects arbitrary number of repositories to donate to
2. Blackbox: subprocess evokes "GetMultiplier" and makes a lot of magic
	- contains multiple subprocesses?
	-> Blackbox: calls subprocess to get all necessary data from database
		- repository trust tag
		- foundations spender limit (reached or not)
			functions:
				- db.FindLastEventMultiplierRepo: Returns last value of Repo with repoid, multiplierAt, unMultiplierAt set
				- GetMultiplierCount: returns per repo amount of active multipliers
	-> Blackbox: with the polled database informations the max amount foundations are paying is calculated
3. Blackbox: Execute HandleFoundationDonation to donate to the repositories
*/

// general structure and functions
/*
- Algorithm: Add function to get and return the multiplier value
- Algorithm: Add function to calculate the multiplier value per repo - loop?
*/
