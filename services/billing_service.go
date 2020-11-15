package services

import (
	"crypto/sha512"
	"encoding/hex"
	"net/http"
	"strconv"
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
func (s *AkouendyPaymentService) PaymentInit(payment Payment) (checkoutUrl string, initError error) {
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
		"cancel_url":   "",
		"return_url":   "",
		"trans_id":     transaction.Token,
		"hash":         hash,
		"webhook":      "",
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

func Hash512(text string) string {
	hasher := sha512.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}
