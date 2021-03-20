package services

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Akouendy-Payments/beego/models"
	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"
	beego "github.com/beego/beego/v2/server/web"
	"github.com/imroc/req"
)

const (
	baseUrl string = "https://pay.akouendy.com"
	// baseUrl      string = "http://localhost:9009"
	initUrl      string = baseUrl + "/v1/billing/payment/init"
	statusUrl    string = baseUrl + "/v1/billing/payment/status"
	checkoutBase string = baseUrl + "/v1/billing/"
)

type AkouendyPaymentService struct {
	merchantId    string
	merchantToken string
	debug         bool
}

func NewPaymentService() *AkouendyPaymentService {
	id, _ := beego.AppConfig.String("payment-merchant-id")
	token, _ := beego.AppConfig.String("payment-merchant-token")
	debug := beego.AppConfig.DefaultBool("payment-debug", false)
	return &AkouendyPaymentService{merchantId: id, merchantToken: token, debug: debug}
}
func (s *AkouendyPaymentService) PaymentInit(payment Payment, oldTrxToken string, callbackUrl string, returnUrl string) (checkoutUrl string, initError error) {
	var transaction models.BillingTransaction
	o := orm.NewOrm()
	if len(oldTrxToken) > 5 {
		o.QueryTable(new(models.BillingTransaction)).Filter("Token", oldTrxToken).Filter("Status", models.PENDING).One(&transaction)
	} else {
		transaction.Status = models.FAILED
		transaction.Provider = payment.Provider
		transaction.OwnerId = payment.UserId
		transaction.ExternalId = payment.ExternalId
		transaction.Amount = payment.Amount
		transaction.Description = payment.Desc
	}
	transaction.MetaData(payment.UserId)
	logs.Info("===== trax====", transaction)

	str := s.merchantId + "|" + transaction.Token + "|" + strconv.Itoa(transaction.Amount) + "|akouna_matata"
	hash := Hash512(str)

	response := PlatformResponse{}
	var header = make(http.Header)
	req.SetTimeout(50 * time.Second)
	req.Debug = s.debug
	header.Set("Content-Type", "application/x-www-form-urlencoded")
	param := req.Param{
		"total_amount": transaction.Amount,
		"description":  transaction.Description,
		"merchant_id":  s.merchantId,
		"cancel_url":   returnUrl,
		"return_url":   returnUrl,
		"trans_id":     transaction.Token,
		"hash":         hash,
		"webhook":      callbackUrl,
	}
	r, err := req.Post(initUrl, param, header)
	statusCode := r.Response().StatusCode
	if err == nil {
		if statusCode == http.StatusOK {
			r.ToJSON(&response)
			transaction.Status = models.PENDING
			transaction.TransactionId = response.Token

			initError := errors.New("Insert or Update error")
			if len(oldTrxToken) > 5 {
				_, initError = o.Update(&transaction, "Token", "Updated")
			} else {
				_, initError = o.Insert(&transaction)
			}

			if initError == nil {
				checkoutUrl = checkoutBase + transaction.Provider + "/" + response.Token
			}

		}
	} else {
		logs.Info("== AkouendyPaymentService str ==== ", err)
	}
	return
}

func (s *AkouendyPaymentService) ValidatePayment(check PaymentCheck) (token string, paymentError error) {
	var status models.TransactionStatus = models.FAILED
	var transaction models.BillingTransaction
	tokens := strings.Split(check.RefCmd, "_")
	token = tokens[0]
	o := orm.NewOrm()
	err := o.QueryTable(new(models.BillingTransaction)).Filter("token", token).One(&transaction)
	if err != orm.ErrNoRows {
		str := s.merchantToken + "|" + check.RefCmd + "|" + strconv.Itoa(check.Status)
		hash := Hash512(str)
		//compare the hash received and the calculated one
		if hash == check.Hash && check.Status == 200 {
			status = models.SUCCESS
			o.QueryTable(new(models.BillingAccount)).Filter("owner_id", transaction.OwnerId).Update(orm.Params{
				"balance": orm.ColValue(orm.ColAdd, transaction.Amount),
			})
		} else {
			paymentError = errors.New("Hash check failed")
		}
		transaction.Status = status
		if _, err := o.Update(&transaction); err != nil {
			paymentError = errors.New("Update transaction failed")
		}

	} else {
		paymentError = orm.ErrNoRows
		logs.Error("Settings row  not found, token :", token)
	}
	return
}
