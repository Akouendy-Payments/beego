package services

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Akouendy/akouendy_payments/models"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"github.com/imroc/req"
)

const (
	//baseUrl string = "https://pay.akouendy.com"
	baseUrl      string = "http://localhost:9009"
	initUrl      string = baseUrl + "/v1/billing/payment/init"
	statusUrl    string = baseUrl + "/v1/billing/payment/status"
	checkoutBase string = baseUrl + "/v1/billing/"
)

type AkouendyPaymentService struct {
	merchantId    string
	merchantToken string
}

func NewPaymentService() *AkouendyPaymentService {
	id := beego.AppConfig.String("payment-merchant-id")
	token := beego.AppConfig.String("payment-merchant-token")
	return &AkouendyPaymentService{merchantId: id, merchantToken: token}
}
func (s *AkouendyPaymentService) PaymentInit(payment Payment, callbackUrl string, returnUrl string) (checkoutUrl string, initError error) {
	logs.Info("== AkouendyPaymentService merchantId ==== " + s.merchantId)
	logs.Info("== AkouendyPaymentService merchantToken ==== " + s.merchantToken)
	transaction := models.BillingTransaction{}
	transaction.MetaData(payment.UserId)
	transaction.Status = models.FAILED
	transaction.Provider = payment.Provider
	transaction.OwnerId = payment.UserId

	str := s.merchantId + "|" + transaction.Token + "|" + strconv.Itoa(payment.Amount) + "|akouna_matata"
	hash := Hash512(str)

	logs.Info("== AkouendyPaymentService str ==== " + str)
	logs.Info("== AkouendyPaymentService hash ==== " + hash)

	response := PlatformResponse{}
	var header = make(http.Header)
	req.SetTimeout(50 * time.Second)
	req.Debug = true
	header.Set("Content-Type", "application/x-www-form-urlencoded")
	param := req.Param{
		"total_amount": payment.Amount,
		"description":  payment.Desc,
		"merchant_id":  s.merchantId,
		"cancel_url":   returnUrl,
		"return_url":   returnUrl,
		"trans_id":     transaction.Token,
		"hash":         hash,
		"webhook":      callbackUrl,
	}
	r, err := req.Post(initUrl, param, header)
	statusCode := r.Response().StatusCode
	logs.Info("== AkouendyPaymentService statusCode ==== ", statusCode)

	if err == nil {
		if statusCode == http.StatusOK {
			r.ToJSON(&response)
			transaction.Status = models.PENDING
			transaction.TransactionId = response.Token
			transaction.Amount = payment.Amount
			transaction.Description = payment.Desc
			o := orm.NewOrm()
			_, initError := o.Insert(&transaction)
			if initError == nil {
				checkoutUrl = checkoutBase + payment.Provider + "/" + response.Token
			}

		}
	} else {
		logs.Info("== AkouendyPaymentService str ==== ", err)
	}
	return
}

func (s *AkouendyPaymentService) ValidatePayment(check PaymentCheck) (paymentError error) {
	var status models.TransactionStatus = models.FAILED
	var transaction models.BillingTransaction
	logs.Info("== ValidatePayment check ==== ", check)
	token := strings.Split(check.RefCmd, "_")
	o := orm.NewOrm()
	err := o.QueryTable(new(models.BillingTransaction)).Filter("token", token).One(&transaction)
	logs.Info("===transaction", transaction)
	logs.Info("=== token", token)
	if err != orm.ErrNoRows {
		str := s.merchantToken + "|" + check.RefCmd + "|" + strconv.Itoa(check.Status)
		hash := Hash512(str)
		//compare the hash received and the calculated one
		if hash == check.Hash && check.Status == 200 {
			status = models.SUCCESS
			o.QueryTable(new(models.Billing)).Filter("owner_id", transaction.OwnerId).Update(orm.Params{
				"balance": orm.ColValue(orm.ColAdd, transaction.Amount),
			})
		} else {
			paymentError = errors.New("Hash check failed")
			logs.Error("From request :", check.Hash)
			logs.Error("Calculated :", hash)
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
