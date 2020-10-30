package main

import (
	"encoding/json"
	"errors"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/customer"
	"github.com/stripe/stripe-go/v72/paymentmethod"
	"github.com/stripe/stripe-go/v72/sub"
	"github.com/stripe/stripe-go/v72/webhook"
	"io/ioutil"
	"log"
	"net/http"
)

func CreateStripeCustomer(user User) (*User, error) {
	params := &stripe.CustomerParams{
		Email: stripe.String(user.Email),

	}
	params.AddMetadata("uid", user.ID)
	c, err := customer.New(params)
	if err != nil{
		return nil, err
	}
	user.StripeId.String = c.ID
	user.StripeId.Valid = true

	// add stripe id to DB
	userErr := UpdateUser(&user)
	if userErr != nil {
		return  nil, userErr
	}
	return &user, nil
}

func CreateSubscription(user User, plan string, paymentMethod string) (*stripe.Subscription, error){
	if !user.StripeId.Valid {
		log.Print("error in createSubscription: user has no stripeID")
		return nil, errors.New("can not create subscription for user without stripeID")
	}
	paymentParams := &stripe.PaymentMethodAttachParams{
		Customer: stripe.String(user.StripeId.String),
	}
	p, err := paymentmethod.Attach(paymentMethod, paymentParams)
	if err != nil {
		log.Printf("Could not set payment method for user %s: %v", user.ID, err)
	}
	subParams := &stripe.SubscriptionParams{
		Customer: stripe.String(user.StripeId.String),
		DefaultPaymentMethod: stripe.String(p.ID),
		Items: []*stripe.SubscriptionItemsParams{
			{
				Price: stripe.String(plan),
			},
		},
	}
	s, sErr := sub.New(subParams)
	if sErr != nil {
		log.Printf("could not create subscription: %v", sErr)
		return nil, sErr
	}
	log.Print("sub created")
	invoice := s.LatestInvoice
	paymentIntent := invoice.PaymentIntent

	if paymentIntent != nil && paymentIntent.Status == "succeeded" {
		log.Print("in if statement status succeeded")
		user.Subscription.String = s.ID
		user.Subscription.Valid = true
		err := UpdateUser(&user)
		if err != nil {
			return nil, err
		}
	}
	return s, nil
}

func StripeWebhook (w http.ResponseWriter, req *http.Request) {
	const MaxBodyBytes = int64(65536)
	req.Body = http.MaxBytesReader(w, req.Body, MaxBodyBytes)
	payload, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Printf("Error reading request body: %v\n", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	event, err := webhook.ConstructEvent(payload, req.Header.Get("Stripe-Signature"), "whsec_4sKlkahWOOeyImUNXxSPBwNJejj2yu9Y")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest) // Return a 400 error on a bad signature
		log.Printf("Error evaluating signed webhook request: %v", err)
		return
	}
	// Unmarshal the event data into an appropriate struct depending on its Type
	switch event.Type {
	case "customer.subscription.created":
		var subscription stripe.Subscription
		err := json.Unmarshal(event.Data.Raw, &subscription)
		if err != nil {
			log.Printf("Error parsing webhook JSON: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		c := subscription.Customer
		cfull, err := customer.Get(c.ID, nil)
		if err != nil {
			log.Printf("customer.subscription.created: Could not retrieve stripe customer %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		user, err := FindUserByID(cfull.Metadata["uid"])
		if err != nil {
			log.Printf("customer.subscription.created: No matching user found %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		user.Subscription.String = subscription.ID
		user.Subscription.Valid = true
		err = UpdateUser(user)
		if err != nil {
			log.Printf("Could not update user %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

	case "customer.subscription.deleted":
		var subscription stripe.Subscription
		err := json.Unmarshal(event.Data.Raw, &subscription)
		if err != nil {
			log.Printf("Error parsing webhook JSON: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		c := subscription.Customer
		cfull, err := customer.Get(c.ID, nil)
		if err != nil {
			log.Printf("customer.subscription.deleted: Could not retrieve stripe customer %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		user, err := FindUserByID(cfull.Metadata["uid"])
		if err != nil {
			log.Printf("customer.subscription.deleted: No matching user found %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		user.Subscription.String = ""
		user.Subscription.Valid = false
		err = UpdateUser(user)
		if err != nil {
			log.Printf("Could not update user %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
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

	// ... handle other event types
	default:
		log.Printf( "Unhandled event type: %s\n", event.Type)
	}

	w.WriteHeader(http.StatusOK)
}