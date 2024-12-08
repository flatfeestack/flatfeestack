package api

import (
	"testing"
)

//	func SetupAnalysisTestServer(t *testing.T) *httptest.Server {
//		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//			switch r.URL.Path {
//			case "/analyze":
//				var request db.AnalysisRequest
//				err := json.NewDecoder(r.Body).Decode(&request)
//				require.Nil(t, err)
//
//				err = json.NewEncoder(w).Encode(client.AnalysisResponse2{RequestId: request.Id})
//				require.Nil(t, err)
//			default:
//				http.NotFound(w, r)
//			}
//		}))
//		return server
//	}
//
//	func insertTestUser(t *testing.T, email string) *db.UserDetail {
//		u := db.User{
//			Id:    uuid.New(),
//			Email: email,
//		}
//		ud := db.UserDetail{
//			User:     u,
//			StripeId: util.StringPointer("strip-id"),
//		}
//
//		err := db.InsertUser(&ud)
//		assert.Nil(t, err)
//		u2, err := db.FindUserById(u.Id)
//		assert.Nil(t, err)
//		return u2
//	}
//
//	func insertPayInEvent(t *testing.T, externalId uuid.UUID, userId uuid.UUID, status string, currency string, amount int64, seats int64, freq int64) *db.PayInEvent {
//		ub := db.PayInEvent{
//			Id:         uuid.New(),
//			ExternalId: externalId,
//			UserId:     userId,
//			Balance:    big.NewInt(amount),
//			Status:     status,
//			Currency:   currency,
//			Seats:      seats,
//			Freq:       freq,
//			CreatedAt:  time.Time{},
//		}
//		err := db.InsertPayInEvent(ub)
//		assert.Nil(t, err)
//		return &ub
//	}
func TestMultiplierCalculation(t *testing.T) {
	//db.SetupTestData()
	//defer db.TeardownTestData()

	/*
		Get Fixed Testdata
			- some repos with sponsors
			- some repos without sponsors
			- calculate how much they should get
	*/

}

// Function: TestMultiplierCalculationOneRepo
/*
test case: one repo, trusted with sponsor
result: no calculation necessary

test case: one repo, untrusted with sponsor
result: no calculation necessary

*/

// Function: TestMultiplierCalculationTwoRepos
/*
test case: two repos
	- both not trusted
	- one sponsored

*/

// Function: TestMultiplierCalculationN_NumberRepos
/*
test case:
	- n repos
	- none trusted
result:
	- no calculation necessary

test case:
	- n repos
	- i repos trusted (fixed)
result:
	- n - i foundations are donating
	-

test case: n repos
	- n repos trusted (randomized)

*/
