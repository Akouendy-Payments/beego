package services

type Payment struct {
	UserId      string
	Amount      int
	Provider    string
	Desc        string
	CheckoutUrl string
}

type PaymentCheck struct {
	RefCmd string
	Status int
	Hash   string
}
