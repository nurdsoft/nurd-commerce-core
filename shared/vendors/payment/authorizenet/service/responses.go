package service

// BaseResponse represents the common response structure from Authorize.net
type BaseResponse struct {
	Messages Messages `json:"messages"`
}

type Messages struct {
	ResultCode string    `json:"resultCode"`
	Message    []Message `json:"message"`
}

type Message struct {
	Code string `json:"code"`
	Text string `json:"text"`
}

type CreateCustomerProfileResponse struct {
	BaseResponse
	CustomerProfileID string `json:"customerProfileId"`
}

type CreateCustomerPaymentProfileResponse struct {
	BaseResponse
	CustomerProfileID        string `json:"customerProfileId"`
	CustomerPaymentProfileID string `json:"customerPaymentProfileId"`
}

type GetCustomerProfileResponse struct {
	BaseResponse
	Profile ProfileResponse `json:"profile"`
}

type ProfileResponse struct {
	PaymentProfiles []PaymentProfileResponse `json:"paymentProfiles"`
}

type PaymentProfileResponse struct {
	CustomerPaymentProfileID string          `json:"customerPaymentProfileId"`
	Payment                  PaymentResponse `json:"payment"`
}

type PaymentResponse struct {
	CreditCard CreditCardResponse `json:"creditCard"`
}

type CreditCardResponse struct {
	CardNumber     string `json:"cardNumber"`
	CardType       string `json:"cardType"`
	ExpirationDate string `json:"expirationDate"`
}

type CreateTransactionResponse struct {
	BaseResponse
	TransactionResponse TransactionResponse `json:"transactionResponse"`
}

type TransactionResponse struct {
	ResponseCode  string    `json:"responseCode"`
	TransID       string    `json:"transId"`
	AccountNumber string    `json:"accountNumber"`
	AccountType   string    `json:"accountType"`
	Messages      []Message `json:"messages,omitempty"`
	Errors        []Error   `json:"errors,omitempty"`
}

type Error struct {
	ErrorCode string `json:"errorCode"`
	ErrorText string `json:"errorText"`
}
