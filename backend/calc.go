package main

import (
	api2 "backend/api"
	"backend/client"
	"backend/db"
	"backend/util"
	"fmt"
	"log/slog"
	"math/big"
	"time"

	"github.com/google/uuid"
)

type CalcHandler struct {
	ac *client.AnalysisClient
	ec *client.EmailClient
}

func NewCalcHandler(ac *client.AnalysisClient, ec *client.EmailClient) *CalcHandler {
	return &CalcHandler{ac, ec}
}

func (c *CalcHandler) HourlyRunner(now time.Time) error {
	//find repos that have an analysis older than 2 days
	a, err := db.FindAllLatestAnalysisRequest(now.AddDate(0, 0, -1))
	if err != nil {
		return err
	}
	slog.Info("Start hourly analysis check",
		slog.Int("len", len(a)))

	nr := 0
	for _, v := range a {
		err := c.ac.RequestAnalysis(v.RepoId, v.GitUrl)
		if err != nil {
			slog.Warn("analysis request failed",
				slog.Any("error", err))
		} else {
			nr++
		}
	}

	slog.Info("Hourly runner processed",
		slog.Int("nr", nr))
	return nil
}

func (c *CalcHandler) DailyRunner(now time.Time) error {
	yesterdayStop := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	yesterdayStart := yesterdayStop.AddDate(0, 0, -1)

	slog.Info("Start daily runner",
		slog.Any("time-start", yesterdayStart),
		slog.Any("time-stop", yesterdayStop))

	sponsorResults, err := db.FindSponsorsBetween(yesterdayStart, yesterdayStop)
	if err != nil {
		return err
	}

	nr := 0
	for key, _ := range sponsorResults {
		if len(sponsorResults[key].RepoIds) > 0 {
			err = c.calcContribution(sponsorResults[key].UserId, sponsorResults[key].RepoIds, yesterdayStart)
			nr++
			if err != nil {
				return err
			}
			allFoundationsPerUser, parts, err := db.GetAllFoundationsSupportingRepos(sponsorResults[key].RepoIds)
			if err != nil {
				return err
			}
			slog.Info("Parts for Multiplier",
				slog.Int("parts", parts))
			if len(allFoundationsPerUser) > 0 {
				err = c.calcMultiplier(sponsorResults[key].UserId, parts, yesterdayStart)
				if err != nil {
					return err
				}
			}
		}
	}

	slog.Info("Daily runner inserted",
		slog.Int("nr", nr))

	//aggregate marketing emails
	ms, err := db.FindMarketingEmails()
	for _, v := range ms {
		if err != nil {
			return err
		}
		repoNames := []string{}
		//TODO: fetch repo names
		err = c.ec.SendMarketingEmail(v.Email, v.Balances, repoNames)
	}

	return nil
}

func (c *CalcHandler) calcMultiplier(uid uuid.UUID, parts int, yesterdayStart time.Time) error {
	currentSponsorDonations, err := db.GetUserDonationRepos(uid, yesterdayStart, false)
	if err != nil {
		return err
	}

	err = calcAndDeductFoundation(currentSponsorDonations, parts, yesterdayStart, false)
	if err != nil {
		return err
	}

	futureSponsorDonations, err := db.GetUserDonationRepos(uid, yesterdayStart, true)
	if err != nil {
		return err
	}

	err = calcAndDeductFoundation(futureSponsorDonations, parts, yesterdayStart, true)
	if err != nil {
		return err
	}

	return nil
}

func calcAndDeductFoundation(sponsorDonations map[uuid.UUID][]db.UserDonationRepo, parts int, yesterdayStart time.Time, futureContribution bool) error {
	for _, currencyBlock := range sponsorDonations {
		for _, block := range currencyBlock {
			if len(block.TrustedRepoSelected) > 0 {
				allRepos := append(block.TrustedRepoSelected, block.UntrustedRepoSelected...)

				amountPerRepo := new(big.Int).Div(&block.SponsorAmount, big.NewInt(int64(len(allRepos))))

				pool := new(big.Int).Mul(amountPerRepo, big.NewInt(int64(len(block.TrustedRepoSelected))))

				var payoutlimit *big.Float
				if futureContribution {
					payoutlimit = new(big.Float).SetInt(amountPerRepo)
				} else {
					payoutlimit = new(big.Float).Mul(new(big.Float).SetInt(amountPerRepo), big.NewFloat(0.9))
				}
				amountPerPart := new(big.Int).Quo(pool, big.NewInt(int64(parts)))

				err := doDeductFoundation(allRepos, yesterdayStart, block.Currency, amountPerPart, payoutlimit, futureContribution)
				return err
			}
		}
	}
	return nil
}

func (c *CalcHandler) calcContribution(uid uuid.UUID, rids []uuid.UUID, yesterdayStart time.Time) error {
	u, err := db.FindUserById(uid)
	if err != nil {
		return fmt.Errorf("cannot find user %v", err)
	}
	//first check if the sponsor has enough funds
	if u.InvitedId != nil {
		u1, err := db.FindUserById(*u.InvitedId)
		if err != nil {
			return fmt.Errorf("cannot find invited user %v", err)
		}
		slog.Info("User sponsored by supports",
			slog.String("email", u.Email),
			slog.String("email", u1.Email),
			slog.Int("len(rids)", len(rids)))
		return c.calcAndDeduct(u1, rids, yesterdayStart, u)
	}
	//TODO: also notify the not only the parent of insufficient funds
	slog.Debug("User supports repos",
		slog.String("email", u.Email),
		slog.Int("len(rids)", len(rids)))
	return c.calcAndDeduct(u, rids, yesterdayStart, nil)
}

func (c *CalcHandler) calcAndDeduct(u *db.UserDetail, rids []uuid.UUID, yesterdayStart time.Time, uOrig *db.UserDetail) error {
	currency, freq, distributeDeduct, distributeAdd, deductFutureContribution, err := calcShare(u.Id, int64(len(rids)))
	if err != nil {
		return fmt.Errorf("cannot calc share %v", err)
	}

	if freq <= 1 {
		slog.Info("1 day or less left, top up!",
			slog.String("email", u.Email),
			slog.String("userId", u.Id.String()))
		c.reminderTopUp(*u, uOrig)
	}

	if freq > 0 {
		slog.Info("User will support these repos",
			slog.String("email", u.Email),
			slog.String("userId", u.Id.String()),
			slog.Any("rids", rids))
		err = doDeduct(u.Id, rids, yesterdayStart, currency, distributeDeduct, distributeAdd, deductFutureContribution)
		return err
	} else {
		slog.Debug("User is out of funds",
			slog.String("userId", u.Id.String()))
	}
	return nil
}

func doDeductFoundation(rids []uuid.UUID, yesterdayStart time.Time, currency string, amountPerPart *big.Int, payoutlimit *big.Float, futureContribution bool) error {
	for _, rid := range rids {
		// Get contributor weights
		uidInMap, uidNotInMap, total, err := getContributorWeights(rid)
		if err != nil {
			return err
		}
		if uidInMap == nil && uidNotInMap == nil {
			continue
		}

		var amountFoundation *big.Float
		foundations, err := db.GetValidatedFoundationsSupportingRepo(rid, currency, yesterdayStart)
		if err != nil {
			return err
		}

		foundationCount := big.NewInt(int64(len(foundations)))
		totalAmountForRepo := new(big.Float).Mul(new(big.Float).SetInt(foundationCount), new(big.Float).SetInt(amountPerPart))

		if totalAmountForRepo.Cmp(payoutlimit) > 0 {
			amountFoundation = new(big.Float).Quo(payoutlimit, new(big.Float).SetInt(foundationCount))
		} else {
			amountFoundation = new(big.Float).SetInt(amountPerPart)
		}

		amountFoundationIntToCheck := new(big.Int)
		amountFoundation.Int(amountFoundationIntToCheck)

		for _, foundation := range foundations {
			amountFoundationToPay, err := db.CheckDailyLimitStillAdheredTo(&foundation, amountFoundationIntToCheck, currency, yesterdayStart)
			if err != nil {
				return err
			}

			slog.Error("after daily limit check\n",
				slog.Any("amount", amountFoundationToPay))

			if amountFoundationToPay.Cmp(big.NewInt(-1)) == 0 {
				continue
			}

			amountFoundationToPay, err = db.CheckFondsAmountEnough(&foundation, amountFoundationToPay, currency)
			if err != nil {
				return err
			}

			slog.Error("after fonds check\n",
				slog.Any("amount", amountFoundationToPay))

			if amountFoundationToPay.Cmp(big.NewInt(-1)) == 0 {
				continue
			}

			for email, w := range uidNotInMap {
				newTotal := total + w
				amount := calcSharePerUser(amountFoundationToPay, w, newTotal)
				slog.Info("Unclaimed / not in map",
					slog.String("userId", foundation.Id.String()),
					slog.String("add", amountFoundationToPay.String()),
					slog.Float64("weight", w),
					slog.Float64("total", newTotal),
					slog.String("amount", amount.String()))
				id := uuid.New()
				err = db.InsertUnclaimed(id, email, rid, amount, currency, yesterdayStart, util.TimeNow())
				if err != nil {
					slog.Error("insertUnclaimed failed: %v, %v\n",
						slog.String("email", email),
						slog.Any("error", err))
				}
			}

			if futureContribution {
				// TODO: to perstist the history of the future cintribution DB table for foudnations, there should be an separate DB table with own unique fields
				slog.Info("Unclaimed / deducted",
					slog.String("rid", rid.String()),
					slog.Float64("total", total),
					slog.String("deduct", amountFoundationToPay.String()))
				err = db.InsertOrUpdateFutureContribution(foundation.Id, rid, amountFoundationToPay, currency, yesterdayStart, util.TimeNow(), true)
				if err != nil {
					return err
				}
			} else {
				mFut, err := db.FindSumFutureSponsorsFromFoundation(foundation.Id)
				if err != nil {
					return err
				}

				distributeFutureAdd := amountFoundationToPay
				var deductFutureContribution *big.Int
				var distributable *big.Int
				if mFut[currency] != nil {
					distributeFutureAdd = new(big.Int).Div(mFut[currency], big.NewInt(int64(len(rids))))
					deductFutureContribution = new(big.Int).Neg(distributeFutureAdd)
				}

				if deductFutureContribution != nil {
					err = db.InsertFutureContribution(foundation.Id, rid, deductFutureContribution, currency, yesterdayStart, util.TimeNow(), true)
					if err != nil {
						return err
					}

					slog.Info("futureContributionTest2",
						slog.Any("yesterdayStart", yesterdayStart),
						slog.String("amountFoundationToPay", deductFutureContribution.String()))

					distributable = new(big.Int).Add(distributeFutureAdd, amountFoundationToPay)
				} else {
					distributable = distributeFutureAdd
				}

				for contributorUserId, w := range uidInMap {
					amount := calcSharePerUser(distributable, w, total)
					slog.Info("Claim",
						slog.String("userId", contributorUserId.String()),
						slog.String("rid", rid.String()),
						slog.String("add", distributable.String()),
						slog.Float64("weight", w),
						slog.Float64("total", total),
						slog.String("amount", amount.String()))
					err = db.InsertContribution(foundation.Id, contributorUserId, rid, amount, currency, yesterdayStart, util.TimeNow(), true)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func doDeduct(uid uuid.UUID, rids []uuid.UUID, yesterdayStart time.Time, currency string, distributeDeduct *big.Int, distributeAdd *big.Int, deductFutureContribution *big.Int) error {
	for _, rid := range rids {
		// Get contributor weights
		uidInMap, uidNotInMap, total, err := getContributorWeights(rid)
		if err != nil {
			return err
		}
		if uidInMap == nil && uidNotInMap == nil {
			continue
		}

		for email, w := range uidNotInMap {
			//pretend that this user is also part of the total, which he is not, but we want
			//to show him what his/her share would be
			newTotal := total + w
			amount := calcSharePerUser(distributeAdd, w, newTotal)
			slog.Info("Unclaimed / not in map",
				slog.String("userId", uid.String()),
				slog.String("add", distributeAdd.String()),
				slog.Float64("weight", w),
				slog.Float64("total", newTotal),
				slog.String("amount", amount.String()))
			id := uuid.New()
			err = db.InsertUnclaimed(id, email, rid, amount, currency, yesterdayStart, util.TimeNow())
			if err != nil {
				slog.Error("insertUnclaimed failed: %v, %v\n",
					slog.String("email", email),
					slog.Any("error", err))
			}
		}

		if len(uidInMap) == 0 {
			//no contribution park the sponsoring separately
			slog.Info("Unclaimed / deducted",
				slog.String("rid", rid.String()),
				slog.String("add", distributeAdd.String()),
				slog.Float64("total", total),
				slog.String("deduct", distributeDeduct.String()))
			err = db.InsertFutureContribution(uid, rid, distributeDeduct, currency, yesterdayStart, util.TimeNow(), false)
			if err != nil {
				return err
			}
		} else {
			var distributable *big.Int

			if deductFutureContribution != nil {
				err = db.InsertFutureContribution(uid, rid, deductFutureContribution, currency, yesterdayStart, util.TimeNow(), false)
				if err != nil {
					return err
				}

				distributable = new(big.Int).Add(distributeAdd, distributeDeduct)
			} else {
				distributable = distributeAdd
			}

			for contributorUserId, w := range uidInMap {
				amount := calcSharePerUser(distributable, w, total)
				slog.Info("Claim",
					slog.String("userId", contributorUserId.String()),
					slog.String("rid", rid.String()),
					slog.String("add", distributable.String()),
					slog.Float64("weight", w),
					slog.Float64("total", total),
					slog.String("amount", amount.String()))
				err = db.InsertContribution(uid, contributorUserId, rid, amount, currency, yesterdayStart, util.TimeNow(), false)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func getContributorWeights(rid uuid.UUID) (map[uuid.UUID]float64, map[string]float64, float64, error) {
	a, err := db.FindLatestAnalysisRequest(rid)
	if err != nil {
		return nil, nil, 0, err
	}
	if a == nil {
		return nil, nil, 0, nil
	}

	ars, err := db.FindAnalysisResults(a.Id)
	if err != nil {
		return nil, nil, 0, err
	}

	uidInMap := map[uuid.UUID]float64{}
	uidNotInMap := map[string]float64{}
	total := 0.0

	for _, ar := range ars {
		uidGit, err := db.FindUserByGitEmail(ar.GitEmail)
		if err != nil {
			return nil, nil, 0, err
		}
		if uidGit != nil {
			uidInMap[*uidGit] += ar.Weight
			total += ar.Weight
		} else {
			uidNotInMap[ar.GitEmail] += ar.Weight
		}
	}

	return uidInMap, uidNotInMap, total, nil
}

func calcSharePerUser(distributeAdd *big.Int, v float64, total float64) *big.Int {
	distributeAddF := new(big.Float).SetInt(distributeAdd)
	amountF := new(big.Float).Mul(big.NewFloat(v), distributeAddF)
	amountF2 := new(big.Float).Quo(amountF, big.NewFloat(total))
	amount := new(big.Int)
	amountF2.Int(amount)
	return amount
}

func calcShare(userId uuid.UUID, repoLen int64) (string, int64, *big.Int, *big.Int, *big.Int, error) {
	//mAdd is what the user paid in the current cycle
	mAdd, err := db.FindSumPaymentByCurrency(userId, db.PayInSuccess)
	if err != nil {
		return "", 0, nil, nil, nil, fmt.Errorf("cannot find sum user balance %v", err)
	}

	//either the user spent it on a repo that does not have any devs who can claim
	mFut, err := db.FindSumFutureSponsors(userId)
	if err != nil {
		return "", 0, nil, nil, nil, fmt.Errorf("cannot find sum user balance %v", err)
	}

	//or the user spent it on for a repo with a dev who can claim
	mSub, err := db.FindSumDailySponsors(userId)
	if err != nil {
		return "", 0, nil, nil, nil, fmt.Errorf("cannot find sum daily balance %v", err)
	}

	currency, freq, s, err := api2.StrategyDeductMax(userId, mAdd, mSub, mFut)

	if s == nil {
		return currency, freq, nil, nil, nil, nil
	}
	//split the contribution among the repos
	distributeDeduct := new(big.Int).Div(s, big.NewInt(repoLen))
	distributeFutureAdd := distributeDeduct
	var deductFutureContribution *big.Int
	if mFut[currency] != nil {
		distributeFutureAdd = new(big.Int).Div(mFut[currency], big.NewInt(repoLen))
		//if we distribute more, we need to deduct this from the future balances
		deductFutureContribution = new(big.Int).Neg(distributeFutureAdd)
	}
	slog.Info("Calculation",
		slog.String("currency", currency),
		slog.Int64("frey", freq),
		slog.String("deduct", distributeDeduct.String()),
		slog.String("add", distributeFutureAdd.String()),
		slog.String("deduct-future", deductFutureContribution.String()))
	return currency, freq, distributeDeduct, distributeFutureAdd, deductFutureContribution, nil
}

func (c *CalcHandler) reminderTopUp(u db.UserDetail, uOrig *db.UserDetail) error {

	//check if user has stripe
	if u.StripeId != nil && u.PaymentMethod != nil {
		err := api2.StripePaymentRecurring(u)
		if err != nil {
			return err
		}

		err = c.ec.SendStripeTopUp(u)
		if err != nil {
			return err
		}
	} else {
		//No stripe, just send email
		isSponsor := uOrig != nil
		if isSponsor {
			err := c.ec.SendTopUpSponsor(u)
			if err != nil {
				return err
			}
		} else {
			if u.InvitedId != nil {
				err := c.ec.SendTopUpInvited(u)
				if err != nil {
					return err
				}
			} else {
				err := c.ec.SendTopUpOther(u)
				if err != nil {
					return err
				}
			}
		}
	}

	slog.Info("TOPUP, you are running out of credit",
		slog.Any("user", u))
	return nil
}
