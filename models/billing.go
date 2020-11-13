package models

import (
	"github.com/astaxie/beego/orm"
	"time"
)

type BillingAccount struct {
	BaseModel
	Balance int
	Expire time.Time `orm:"auto_now;type(datetime)"`
}

func init() {
	orm.RegisterModel(new(BillingAccount))
}
