package services

type Payment struct {
	UserId      string
	Amount      int
	Provider    string
	Desc        string
	CheckoutUrl string
}
