package models

import (
	"github.com/astaxie/beego/orm"
	"time"
)

type BillingHistory struct {
	BaseModel
	Balance int
	Start time.Time `orm:"auto_now;type(datetime)"`
	End time.Time `orm:"auto_now;type(datetime)"`
	Description string
	NextUpdate time.Time `orm:"auto_now;type(datetime)"`
}

func init() {
	orm.RegisterModel(new(BillingAccount))
}
