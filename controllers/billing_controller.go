package controllers

import (
	"akouendy-sms-engine/domain"
	"encoding/json"
	"strings"

	"github.com/akouendy-payments/beego/models"
	"github.com/akouendy-payments/beego/services"
	"github.com/astaxie/beego"
	"github.com/jinzhu/copier"
)

//  BillingController operations for Billing
type BillingController struct {
	AuthorizeController
}

// Post ...
// @Title Post
// @Description create Billing
// @Param   Authorization  header   string      true "Bearer JwtToken"
// @Param	body		body 	models.BillingAccount	false		"body for Billing content"
// @Success 201 {int} models.BillingAccount
// @Failure 403 body is empty
// @router /create [post]
func (c *BillingController) CreateAccount() {
	var v models.BillingAccount
	json.Unmarshal(c.Ctx.Input.RequestBody, &v)
	v.MetaData(userId)
	v.OwnerId = userId
	if _, err := models.AddBilling(&v); err == nil {
		c.Ctx.Output.SetStatus(201)
		c.Data["json"] = v
	} else {
		c.Data["json"] = err.Error()
	}
	c.ServeJSON()
}

// GetAll ...
// @Title Get All
// @Description get Billing
// @Param   Authorization  header   string      true "Bearer JwtToken"
// @Param	fields	query	string	false	"Fields returned. e.g. col1,col2 ..."
// @Param	sortby	query	string	false	"Sorted-by fields. e.g. col1,col2 ..."
// @Param	order	query	string	false	"Order corresponding to each sortby field, if single value, apply to all sortby fields. e.g. desc,asc ..."
// @Param	limit	query	string	false	"Limit the size of result set. Must be an integer"
// @Param	offset	query	string	false	"Start position of result set. Must be an integer"
// @Success 200 {object} models.Billing
// @Failure 403
// @router / [get]
func (c *BillingController) GetAll() {
	var fields []string
	var sortby []string
	var order []string
	var query = make(map[string]string)
	var limit int64 = 10
	var offset int64

	// fields: col1,col2,entity.col3
	if v := c.GetString("fields"); v != "" {
		fields = strings.Split(v, ",")
	}
	// limit: 10 (default is 10)
	if v, err := c.GetInt64("limit"); err == nil {
		limit = v
	}
	// offset: 0 (default is 0)
	if v, err := c.GetInt64("offset"); err == nil {
		offset = v
	}
	// sortby: col1,col2
	if v := c.GetString("sortby"); v != "" {
		sortby = strings.Split(v, ",")
	}
	// order: desc,asc
	if v := c.GetString("order"); v != "" {
		order = strings.Split(v, ",")
	}
	// query: k:v,k:v
	query["OwnerId"] = userId

	l, err := models.GetAllBilling(query, fields, sortby, order, offset, limit)
	if err != nil {
		c.Data["json"] = err.Error()
	} else {
		c.Data["json"] = l
	}
	c.ServeJSON()
}

// Post ...
// @Title Post
// @Description charge Account
// @Param   Authorization  header   string      true "Bearer JwtToken"
// @Param	body		body 	domain.Payment	true		"body for Billing content"
// @Success 201 {int} domain.Payment
// @Failure 403 body is empty
// @router /charge-account [put]
func (c *BillingController) ChargeAccount() {
	var payment services.Payment
	var response domain.Payment
	json.Unmarshal(c.Ctx.Input.RequestBody, &payment)
	payment.UserId = userId
	//TODO: update comment stuff
	//payment.Provider = "orange-money-sn"
	payment.Provider = "sandbox"
	payment.Desc = "Akouendy Sms - Transaction"
	paymentService := services.NewPaymentService()
	callbackUrl := beego.AppConfig.String("baseurl") + "/payment/callback"
	returnUrl := beego.AppConfig.String("front-url")
	checkoutUrl, error := paymentService.PaymentInit(payment, callbackUrl, returnUrl)
	if error == nil {
		copier.Copy(&response, &payment)
		response.CheckoutUrl = checkoutUrl
	}
	c.Data["json"] = response
	c.ServeJSON()
}
