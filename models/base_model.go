package models

import (
	"fmt"
	"time"

	uuid "github.com/satori/go.uuid"
)

type BaseModel struct {
	Created   time.Time `orm:"auto_now_add;type(datetime)"`
	Updated   time.Time `orm:"auto_now;type(datetime)"`
	CreatedBy string
	Token     string
	OwnerId   string
}

// multiple fields index
func (b *BaseModel) TableIndex() [][]string {
	return [][]string{
		[]string{"Token"},
		[]string{"OwnerId"},
		[]string{"Updated"},
	}
}

func (b *BaseModel) MetaData(userId string) {
	u2 := uuid.NewV4()
	b.Token = u2.String()
	fmt.Println(" === my user id===:", userId)
	b.CreatedBy = userId
	b.Updated = time.Now()
}
