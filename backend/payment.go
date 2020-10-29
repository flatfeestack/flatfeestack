package main

import (
	"errors"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/customer"
	"github.com/stripe/stripe-go/v72/sub"
	"github.com/stripe/stripe-go/v72/paymentmethod"
	"log"
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