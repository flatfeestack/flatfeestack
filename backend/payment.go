package main

import (
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/charge"
	"github.com/stripe/stripe-go/v72/customer"
	"github.com/stripe/stripe-go/v72/paymentmethod"
	"github.com/stripe/stripe-go/v72/sub"
	"github.com/stripe/stripe-go/v72/webhook"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type PostSubscriptionBody struct {
	Plan          string `json:"plan"`
	PaymentMethod string `json:"paymentMethod"`
}

// @Summary Create a subscription
// @Tags Repos
// @Param body body PostSubscriptionBody true "Request Body"
// @Accept  json
// @Produce  json
// @Success 200
// @Failure 400
// @Router /backend/payments/subscriptions [post]
func postSubscription(w http.ResponseWriter, r *http.Request, user *User) {
	var body PostSubscriptionBody
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Could not decode json: %v", err)
		return
	}

	s, err := createSubscription(*user, body.Plan, body.PaymentMethod)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "Something went wrong: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(s)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "Could not encode json: %v", err)
		return
	}
}

func createStripeCustomer(user *User) (string, error) {
	params := &stripe.CustomerParams{
		Email: stripe.String(*user.Email),
	}
	params.AddMetadata("uid", user.Id.String())
	c, err := customer.New(params)
	if err != nil {
		return "", err
	}
	return c.ID, nil
}

func createSubscription(user User, plan string, paymentMethod string) (*stripe.Subscription, error) {
	if user.StripeId == nil {
		log.Print("error in createSubscription: user has no stripeID")
		return nil, errors.New("can not stringPointer subscription for user without stripeID")
	}
	if opts.Env == "local" {
		user.Subscription = user.StripeId
		a := "Active"
		user.SubscriptionState = &a
		err := updateUser(&user)
		if err != nil {
			return nil, err
		}
		return &stripe.Subscription{
			APIResource:                   stripe.APIResource{},
			ApplicationFeePercent:         0,
			BillingCycleAnchor:            0,
			BillingThresholds:             nil,
			CancelAt:                      0,
			CancelAtPeriodEnd:             false,
			CanceledAt:                    0,
			CollectionMethod:              "",
			Created:                       0,
			CurrentPeriodEnd:              0,
			CurrentPeriodStart:            0,
			Customer:                      nil,
			DaysUntilDue:                  0,
			DefaultPaymentMethod:          nil,
			DefaultSource:                 nil,
			DefaultTaxRates:               nil,
			Discount:                      nil,
			EndedAt:                       0,
			ID:                            "",
			Items:                         nil,
			LatestInvoice:                 nil,
			Livemode:                      false,
			Metadata:                      nil,
			NextPendingInvoiceItemInvoice: 0,
			Object:                        "",
			OnBehalfOf:                    nil,
			PauseCollection:               stripe.SubscriptionPauseCollection{},
			PendingInvoiceItemInterval:    stripe.SubscriptionPendingInvoiceItemInterval{},
			PendingSetupIntent:            nil,
			PendingUpdate:                 nil,
			Plan:                          nil,
			Quantity:                      0,
			Schedule:                      nil,
			StartDate:                     0,
			Status:                        "",
			TransferData:                  nil,
			TrialEnd:                      0,
			TrialStart:                    0,
		}, nil
	}
	paymentParams := &stripe.PaymentMethodAttachParams{
		Customer: stripe.String(*user.StripeId),
	}
	p, err := paymentmethod.Attach(paymentMethod, paymentParams)
	if err != nil {
		log.Printf("Could not set payment method for user %s: %v", user.Id, err)
	}
	subParams := &stripe.SubscriptionParams{
		Customer:             stripe.String(*user.StripeId),
		DefaultPaymentMethod: stripe.String(p.ID),
		Items: []*stripe.SubscriptionItemsParams{
			{
				Price: stripe.String(plan),
			},
		},
	}
	subParams.AddExpand("latest_invoice.payment_intent")
	s, sErr := sub.New(subParams)
	if sErr != nil {
		log.Printf("could not stringPointer subscription: %v", sErr)
		return nil, sErr
	}
	log.Print("sub created")
	invoice := s.LatestInvoice
	paymentIntent := invoice.PaymentIntent

	if paymentIntent != nil && paymentIntent.Status == "succeeded" {
		log.Print("in if statement status succeeded")
		user.Subscription = &s.ID
		state := "active"
		user.SubscriptionState = &state
		err := updateUser(&user)
		if err != nil {
			return nil, err
		}
	}
	return s, nil
}

// @Summary Stripe Webhook handler
// @Tags Webhooks
// @Param user body interface{} true "Request Body"
// @Accept  json
// @Produce  json
// @Success 200
// @Failure 400
// @Router /backend/hooks/stripe [post]
func stripeWebhook(w http.ResponseWriter, req *http.Request) {
	const MaxBodyBytes = int64(65536)
	req.Body = http.MaxBytesReader(w, req.Body, MaxBodyBytes)
	payload, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Printf("Error reading request body: %v\n", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	event, err := webhook.ConstructEvent(payload, req.Header.Get("Stripe-Signature"), "whsec_kWkkVF7yCS3n2SoWf7XLJe3TbKEpN5f1")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest) // Return a 400 error on a bad signature
		log.Printf("Error evaluating signed webhook request: %v", err)
		return
	}
	// Unmarshal the event data into an appropriate struct depending on its Type
	switch event.Type {
	case "customer.subscription.created", "customer.subscription.updated", "customer.subscription.deleted":
		var subscription stripe.Subscription
		err := json.Unmarshal(event.Data.Raw, &subscription)
		if err != nil {
			log.Printf("Error parsing webhook JSON: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// With 3D secure, the subscription may also be created with status incomplete,
		// because the user first has to verify the payment. In this case, the subscription
		// will be activated in the customer.subscription.updated hook.
		w.WriteHeader(checkSubscription(&subscription))

	case "invoice.payment_failed":
		var invoice stripe.Invoice
		err := json.Unmarshal(event.Data.Raw, &invoice)
		if err != nil {
			log.Printf("Error parsing webhook JSON: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		c := invoice.Customer
		cfull, err := customer.Get(c.ID, nil)
		if err != nil {
			log.Printf("invoice.payment_failed: Could not retrieve stripe customer %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Printf("invoice failed for user %s", cfull.Metadata["uid"])
		// TODO: notify user that he needs to update CC
	case "invoice.payment_succeeded":
		var invoice stripe.Invoice
		err := json.Unmarshal(event.Data.Raw, &invoice)
		if err != nil {
			log.Printf("Error parsing webhook JSON: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		cu := invoice.Customer
		cufull, err := customer.Get(cu.ID, nil)
		if err != nil {
			log.Printf("invoice.payment_succeeded: Could not retrieve stripe customer %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		params := &stripe.ChargeParams{}
		params.AddExpand("balance_transaction")
		ch, ch_err := charge.Get(invoice.Charge.ID, params)
		if ch_err != nil {
			log.Printf("Error retrieving charge %v\n", ch_err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		net := ch.BalanceTransaction.Net
		paid := ch.BalanceTransaction.Amount
		if len(invoice.Lines.Data) != 0 && invoice.Lines.Data[0].Period != nil {
			invoiceData := invoice.Lines.Data[0]
			uuid, err := uuid.Parse(cufull.Metadata["uid"])
			err = insertPayment(&Payment{Uid: uuid, From: time.Unix(invoiceData.Period.Start, 0), To: time.Unix(invoiceData.Period.End, 0), Sub: invoiceData.ID, Amount: net})

			if err != nil {
				log.Printf("invoice.payment_succeeded: Error saving payment %v\n", err)
				// return OK to stripe, as the error is on our side
				w.WriteHeader(http.StatusOK)
				return
			}
		}
		log.Printf("Payment stored to database: customer %s paid %v, flatfeestack received %v \n", invoice.Customer.ID, paid, net)
	// ... handle other event types
	default:
		log.Printf("Unhandled event type: %s\n", event.Type)
		w.WriteHeader(http.StatusOK)
	}

}

// Helpers
func checkSubscription(subscription *stripe.Subscription) int {
	c := subscription.Customer
	cfull, err := customer.Get(c.ID, nil)
	if err != nil {
		log.Printf("customer.subscription.created: Could not retrieve stripe customer %v\n", err)
		return http.StatusBadRequest
	}
	uid, err := uuid.Parse(cfull.Metadata["uid"])
	if err != nil {
		log.Printf("customer.subscription.created: Could not retrieve stripe customer %v\n", err)
		return http.StatusBadRequest
	}
	user, err := findUserById(uid)
	if err != nil {
		log.Printf("customer.subscription.created: No matching user found %v\n", err)
		return http.StatusBadRequest
	}

	active := "active"

	if subscription.Status == "active" {
		user.Subscription = &subscription.ID
		user.SubscriptionState = &active
		err = updateUser(user)
		if err != nil {
			log.Printf("Could not update user %v\n", err)
			return http.StatusInternalServerError
		}
	} else if subscription.Status == "canceled" || subscription.Status == "unpaid" {
		// TODO: Check stripe retry rules and deactivate subscription at some point
		user.SubscriptionState = nil
		user.Subscription = nil
		err = updateUser(user)
		if err != nil {
			log.Printf("Could not update user %v\n", err)
			return http.StatusInternalServerError
		}
	} else {
		// Keep subscriptionId the same, but update the status so the subscription can be reactivated later
		s := string(subscription.Status)
		user.SubscriptionState = &s
		user.Subscription = &subscription.ID
		err = updateUser(user)
		if err != nil {
			log.Printf("Could not update user %v\n", err)
			return http.StatusInternalServerError
		}
	}
	return http.StatusOK
}
