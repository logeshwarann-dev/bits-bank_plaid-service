package db

// type SignUpForm struct {
// 	Email       string `json:"email"`
// 	Password    string `json:"password"`
// 	FirstName   string `json:"firstName"`
// 	LastName    string `json:"lastName"`
// 	Address1    string `json:"address1"`
// 	City        string `json:"city"`
// 	State       string `json:"state"`
// 	PostalCode  string `json:"postalCode"`
// 	DateOfBirth string `json:"dob"`
// 	AadharNo    string `json:"aadharNo"`
// }

// type SignInForm struct {
// 	Email    string `json:"email"`
// 	Password string `json:"password"`
// }

// type BankUser struct {
// 	Email       string `gorm:"primaryKey"`
// 	Password    string `gorm:"not null"`
// 	FirstName   string `gorm:"not null"`
// 	LastName    string `gorm:"not null"`
// 	Address1    string `gorm:"not null"`
// 	City        string `gorm:"not null"`
// 	State       string `gorm:"not null"`
// 	PostalCode  string `gorm:"not null"`
// 	DateOfBirth string `gorm:"not null"`
// 	AadharNo    string `gorm:"unique;not null"`
// }

type PlaidUser struct {
	TrackId          string `gorm:"primaryKey"`
	AccountId        string `gorm:"not null"`
	BankId           string `gorm:"not null"`
	AccessToken      string `gorm:"not null"`
	FundingSourceUrl string `gorm:"not null"`
	ShareableId      string `gorm:"not null"`
	UserId           string `gorm:"not null"`
}

func (PlaidUser) TableName() string {
	return "plaid_users"
}

type Transaction struct {
	TransactionId  string `gorm:"primaryKey"`
	Name           string `gorm:"not null"`
	Amount         string `gorm:"not null"`
	Channel        string `gorm:"not null"`
	Category       string `gorm:"not null"`
	SenderId       string `gorm:"not null"`
	ReceiverId     string `gorm:"not null"`
	SenderBankId   string `gorm:"not null"`
	ReceiverBankId string `gorm:"not null"`
}

func (Transaction) TableName() string {
	return "transactions"
}

// func (s *SignUpForm) ConvertToUser() *BankUser {
// 	return &BankUser{
// 		Email:       s.Email,
// 		Password:    s.Password,
// 		FirstName:   s.FirstName,
// 		LastName:    s.LastName,
// 		Address1:    s.Address1,
// 		City:        s.City,
// 		State:       s.State,
// 		PostalCode:  s.PostalCode,
// 		DateOfBirth: s.DateOfBirth,
// 		AadharNo:    s.AadharNo,
// 	}
// }

// type LoggedInUser struct {
// 	Username    string `json:"username" binding:"required"`
// 	FirstName   string `json:"firstName"`
// 	LastName    string `json:"lastName"`
// 	Address1    string `json:"address1"`
// 	City        string `json:"city"`
// 	State       string `json:"state"`
// 	PostalCode  string `json:"postalCode"`
// 	DateOfBirth string `json:"dob"`
// 	AadharNo    string `json:"aadharNo"`
// }
