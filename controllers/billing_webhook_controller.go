package controllers

import (
	"akouendy-sms-engine/domain"
	"net/http"

	"github.com/akouendy-payments/beego/services"
	"github.com/astaxie/beego"
)

type BillingWebHookController struct {
	beego.Controller
}

func (c *BillingWebHookController) Post() {
	var paymentCheck services.PaymentCheck
	checkService := services.NewPaymentService()
	paymentCheck.RefCmd = c.GetString("REF_CMD")
	paymentCheck.Status, _ = c.GetInt("STATUT")
	paymentCheck.Hash = c.GetString("HASH")
	err := checkService.ValidatePayment(paymentCheck)
	if err == nil {
		response := domain.ApiResponse{}
		response.Code = 200
		c.Data["json"] = response
	} else {
		responseError := domain.ErrorResponse{}
		responseError.Raise(c.Ctx, http.StatusInternalServerError, "Payment failed")
	}

	c.ServeJSON()
}
