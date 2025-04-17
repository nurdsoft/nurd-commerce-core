package entities

type CreateSFUserRequest struct {
	FirstName   string `json:"FirstName"`
	LastName    string `json:"LastName"`
	PersonEmail string `json:"PersonEmail"`
}

type UpdateSFUserRequest struct {
	// The Id field should not be specified in the sobject data. Only use for reference in the URL.
	ID        string `json:"-"`
	FirstName string `json:"FirstName"`
	LastName  string `json:"LastName"`
	Phone     string `json:"Phone"`
}

type CreateSFAddressRequest struct {
	AccountC               string `json:"Account__c,omitempty"`
	ShippingStreetC        string `json:"Shipping_Street__c,omitempty"`
	ShippingCityC          string `json:"Shipping_City__c,omitempty"`
	ShippingStateProvinceC string `json:"Shipping_State_Province__c,omitempty"`
	ShippingCountryC       string `json:"Shipping_Country__c,omitempty"`
	ShippingZipPostalCodeC string `json:"Shipping_Zip_Postal_Code__c,omitempty"`
}

type UpdateSFAddressRequest struct {
	AddressID              string `json:"-"` // The Id field should not be specified in the sobject data. Only use for reference in the URL.
	AccountC               string `json:"Account__c,omitempty"`
	ShippingStreetC        string `json:"Shipping_Street__c,omitempty"`
	ShippingCityC          string `json:"Shipping_City__c,omitempty"`
	ShippingStateProvinceC string `json:"Shipping_State_Province__c,omitempty"`
	ShippingCountryC       string `json:"Shipping_Country__c,omitempty"`
	ShippingZipPostalCodeC string `json:"Shipping_Zip_Postal_Code__c,omitempty"`
}

type CreateSFProductRequest struct {
	Name        string `json:"Name"`
	ProductCode string `json:"ProductCode"`
	Description string `json:"Description"`
	IsActive    bool   `json:"IsActive"`
}

type CreateSFPriceBookEntryRequest struct {
	Pricebook2ID string `json:"Pricebook2Id"`
	Product2ID   string `json:"Product2Id"`
	UnitPrice    int    `json:"UnitPrice"`
	IsActive     bool   `json:"IsActive"`
}
type CreateSFOrderRequest struct {
	EffectiveDate           string `json:"EffectiveDate"`
	AccountID               string `json:"AccountId"`
	Status                  string `json:"Status"`
	BillingStreet           string `json:"BillingStreet"`
	BillingCity             string `json:"BillingCity"`
	BillingState            string `json:"BillingState"`
	BillingPostalCode       string `json:"BillingPostalCode"`
	BillingCountry          string `json:"BillingCountry"`
	ShippingStreet          string `json:"ShippingStreet"`
	ShippingCity            string `json:"ShippingCity"`
	ShippingState           string `json:"ShippingState"`
	ShippingPostalCode      string `json:"ShippingPostalCode"`
	ShippingCountry         string `json:"ShippingCountry"`
	Pricebook2ID            string `json:"Pricebook2Id"`
	ShippingRateC           string `json:"shipping_rate__c"`
	TaxAmountC              string `json:"Tax_Amount__c"`
	ShippingCarrierNameC    string `json:"Shipping_Carrier_Name__c"`
	ShippingCarrierServiceC string `json:"Shipping_Carrier_Service__c"`
	CurrencyC               string `json:"Currency__c"`
	SubTotalC               string `json:"SubTotal__c"`
	TotalC                  string `json:"Total__c"`
	OrderReferenceC         string `json:"Order_Reference__c"`
	OrderCreatedAtC         string `json:"Order_Created_At__c"`
	EstimatedDeliveryDateC  string `json:"Estimated_Delivery_Date__c"`
}

type AddOrderItemRequest struct {
	BatchRequests []BatchRequests `json:"batchRequests"`
}
type OrderItem struct {
	OrderID          string  `json:"OrderId"`
	PricebookEntryID string  `json:"PricebookEntryId"`
	Quantity         int     `json:"Quantity"`
	UnitPrice        float64 `json:"UnitPrice"`
	Description      string  `json:"Description"`
	// TypeC is a custom field in Salesforce to store product variant information
	TypeC string `json:"type__c"`
}
type BatchRequests struct {
	Method    string    `json:"method"`
	URL       string    `json:"url"`
	RichInput OrderItem `json:"richInput"`
}

type UpdateOrderRequest struct {
	OrderId   string `json:"-"`
	AccountID string `json:"AccountId"`
	Status    string `json:"Status"`
}
