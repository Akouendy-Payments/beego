package models

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/astaxie/beego/orm"
)

type TransactionStatus string

const (
	PENDING TransactionStatus = "PENDING"
	FAILED  TransactionStatus = "FAILED"
	SUCCESS TransactionStatus = "SUCCESS"
)

type BillingTransaction struct {
	Id int64 `orm:"auto"`
	BaseModel
	Status        TransactionStatus
	TransactionId string
	Provider      string `orm:"size(128)"`
	Amount        int
	Description   string
}

func init() {
	orm.RegisterModel(new(BillingTransaction))
}

// AddBillingTransaction insert a new BillingTransaction into database and returns
// last inserted Id on success.
func AddBillingTransaction(m *BillingTransaction) (id int64, err error) {
	o := orm.NewOrm()
	id, err = o.Insert(m)
	return
}

// GetBillingTransactionById retrieves BillingTransaction by Id. Returns error if
// Id doesn't exist
func GetBillingTransactionById(id int64) (v *BillingTransaction, err error) {
	o := orm.NewOrm()
	v = &BillingTransaction{Id: id}
	if err = o.QueryTable(new(BillingTransaction)).Filter("Id", id).RelatedSel().One(v); err == nil {
		return v, nil
	}
	return nil, err
}

// GetAllBillingTransaction retrieves all BillingTransaction matches certain condition. Returns empty list if
// no records exist
func GetAllBillingTransaction(query map[string]string, fields []string, sortby []string, order []string,
	offset int64, limit int64) (ml []interface{}, err error) {
	o := orm.NewOrm()
	qs := o.QueryTable(new(BillingTransaction))
	// query k=v
	for k, v := range query {
		// rewrite dot-notation to Object__Attribute
		k = strings.Replace(k, ".", "__", -1)
		qs = qs.Filter(k, v)
	}
	// order by:
	var sortFields []string
	if len(sortby) != 0 {
		if len(sortby) == len(order) {
			// 1) for each sort field, there is an associated order
			for i, v := range sortby {
				orderby := ""
				if order[i] == "desc" {
					orderby = "-" + v
				} else if order[i] == "asc" {
					orderby = v
				} else {
					return nil, errors.New("Error: Invalid order. Must be either [asc|desc]")
				}
				sortFields = append(sortFields, orderby)
			}
			qs = qs.OrderBy(sortFields...)
		} else if len(sortby) != len(order) && len(order) == 1 {
			// 2) there is exactly one order, all the sorted fields will be sorted by this order
			for _, v := range sortby {
				orderby := ""
				if order[0] == "desc" {
					orderby = "-" + v
				} else if order[0] == "asc" {
					orderby = v
				} else {
					return nil, errors.New("Error: Invalid order. Must be either [asc|desc]")
				}
				sortFields = append(sortFields, orderby)
			}
		} else if len(sortby) != len(order) && len(order) != 1 {
			return nil, errors.New("Error: 'sortby', 'order' sizes mismatch or 'order' size is not 1")
		}
	} else {
		if len(order) != 0 {
			return nil, errors.New("Error: unused 'order' fields")
		}
	}

	var l []BillingTransaction
	qs = qs.OrderBy(sortFields...).RelatedSel()
	if _, err = qs.Limit(limit, offset).All(&l, fields...); err == nil {
		if len(fields) == 0 {
			for _, v := range l {
				ml = append(ml, v)
			}
		} else {
			// trim unused fields
			for _, v := range l {
				m := make(map[string]interface{})
				val := reflect.ValueOf(v)
				for _, fname := range fields {
					m[fname] = val.FieldByName(fname).Interface()
				}
				ml = append(ml, m)
			}
		}
		return ml, nil
	}
	return nil, err
}

// UpdateBillingTransaction updates BillingTransaction by Id and returns error if
// the record to be updated doesn't exist
func UpdateBillingTransactionById(m *BillingTransaction) (err error) {
	o := orm.NewOrm()
	v := BillingTransaction{Id: m.Id}
	// ascertain id exists in the database
	if err = o.Read(&v); err == nil {
		var num int64
		if num, err = o.Update(m); err == nil {
			fmt.Println("Number of records updated in database:", num)
		}
	}
	return
}

// DeleteBillingTransaction deletes BillingTransaction by Id and returns error if
// the record to be deleted doesn't exist
func DeleteBillingTransaction(id int64) (err error) {
	o := orm.NewOrm()
	v := BillingTransaction{Id: id}
	// ascertain id exists in the database
	if err = o.Read(&v); err == nil {
		var num int64
		if num, err = o.Delete(&BillingTransaction{Id: id}); err == nil {
			fmt.Println("Number of records deleted in database:", num)
		}
	}
	return
}
