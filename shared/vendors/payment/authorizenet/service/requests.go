package service

type merchantAuthentication struct {
	Name           string `json:"name"`
	TransactionKey string `json:"transactionKey"`
}

type Profile struct {
	MerchantCustomerID string `json:"merchantCustomerId"`
	Description        string `json:"description"`
	Email              string `json:"email"`
}

type CreateCustomerPaymentProfileRequest struct {
	Data CreateCustomerPaymentProfileRequestData `json:"createCustomerPaymentProfileRequest"`
}

type CreateCustomerPaymentProfileRequestData struct {
	MerchantAuthentication merchantAuthentication `json:"merchantAuthentication"`
	CustomerProfileID      string                 `json:"customerProfileId"`
	PaymentProfile         PaymentProfile         `json:"paymentProfile"`
	ValidationMode         string                 `json:"validationMode"`
}

type PaymentProfile struct {
	Payment               Payment `json:"payment"`
	DefaultPaymentProfile bool    `json:"defaultPaymentProfile"`
}

type Payment struct {
	CreditCard CreditCard `json:"creditCard"`
}

type CreditCard struct {
	CardNumber     string `json:"cardNumber"`
	ExpirationDate string `json:"expirationDate"`
}

type CreateCustomerProfileRequest struct {
	Data CreateCustomerProfileRequestData `json:"createCustomerProfileRequest"`
}

type CreateCustomerProfileRequestData struct {
	MerchantAuthentication merchantAuthentication `json:"merchantAuthentication"`
	Profile                Profile                `json:"profile"`
}

type GetCustomerProfileRequestData struct {
	MerchantAuthentication merchantAuthentication `json:"merchantAuthentication"`
	CustomerProfileIID     string                 `json:"customerProfileId"`
	UnmaskExpirationDate   bool                   `json:"unmaskExpirationDate"`
}

type GetCustomerProfileRequest struct {
	Data GetCustomerProfileRequestData `json:"getCustomerProfileRequest"`
}

type CreateTransactionRequest struct {
	Data TransactionRequestData `json:"createTransactionRequest"`
}

type TransactionRequestData struct {
	MerchantAuthentication merchantAuthentication `json:"merchantAuthentication"`
	TransactionRequest     TransactionRequest     `json:"transactionRequest"`
}

type TransactionRequest struct {
	TransactionType string       `json:"transactionType"`
	Amount          string       `json:"amount"`
	Payment         PaymentNonce `json:"payment"`
	BillTo          BillTo       `json:"billTo,omitzero"`
}

type PaymentNonce struct {
	OpaqueData OpaqueData `json:"opaqueData"`
}

type OpaqueData struct {
	DataDescriptor string `json:"dataDescriptor,omitempty"`
	DataValue      string `json:"dataValue,omitempty"`
}

type BillTo struct {
	FirstName string `json:"firstName,omitempty"`
	LastName  string `json:"lastName,omitempty"`
	Address   string `json:"address,omitempty"`
	City      string `json:"city,omitempty"`
	State     string `json:"state,omitempty"`
	Zip       string `json:"zip,omitempty"`
	Country   string `json:"country,omitempty"`
}

type RequestData struct {
	TransactionRequest TransactionRequest `json:"transactionRequest"`
}
