package entities

type XAVResponseWrapper struct {
	XAVResponse XAVResponse `json:"XAVResponse"`
}

type XAVResponse struct {
	Response              Response               `json:"Response"`
	ValidAddressIndicator string                 `json:"ValidAddressIndicator"`
	Candidate             []XAVResponseCandidate `json:"Candidate"`
}

type Response struct {
	ResponseStatus ResponseStatus `json:"ResponseStatus"`
}
type ResponseStatus struct {
	Code        string `json:"Code"`
	Description string `json:"Description"`
}
type XAVResponseCandidate struct {
	AddressKeyFormat AddressKeyFormat `json:"AddressKeyFormat"`
}


type RateResponseWrapper struct {
	RateResponse RateResponse `json:"RateResponse"`
}

type RateResponse struct {
	Response      RateResponseDetail `json:"Response"`
	RatedShipment []RatedShipment    `json:"RatedShipment"`
}

type RateResponseDetail struct {
	ResponseStatus       ResponseStatus  `json:"ResponseStatus"`
	Alert                []ResponseAlert `json:"Alert,omitempty"`
	TransactionReference TransactionRef  `json:"TransactionReference"`
}

type ResponseAlert struct {
	Code        string `json:"Code"`
	Description string `json:"Description"`
}

type TransactionRef struct {
	CustomerContext       string `json:"CustomerContext"`
	TransactionIdentifier string `json:"TransactionIdentifier"`
}

type RatedShipment struct {
	Service               CodeDescription     `json:"Service"`
	Zone                  string              `json:"Zone"`
	RatedShipmentAlert    []ResponseAlert     `json:"RatedShipmentAlert,omitempty"`
	BillingWeight         Weight              `json:"BillingWeight"`
	TransportationCharges MonetaryValue       `json:"TransportationCharges"`
	BaseServiceCharge     MonetaryValue       `json:"BaseServiceCharge"`
	ServiceOptionsCharges MonetaryValue       `json:"ServiceOptionsCharges"`
	TotalCharges          MonetaryValue       `json:"TotalCharges"`
	GuaranteedDelivery    *GuaranteedDelivery `json:"GuaranteedDelivery,omitempty"`
	RatedPackage          []RatedPackage      `json:"RatedPackage"`
}

type Weight struct {
	UnitOfMeasurement CodeDescription `json:"UnitOfMeasurement"`
	Weight            string          `json:"Weight"`
}

type MonetaryValue struct {
	CurrencyCode  string `json:"CurrencyCode"`
	MonetaryValue string `json:"MonetaryValue"`
}

type GuaranteedDelivery struct {
	BusinessDaysInTransit string `json:"BusinessDaysInTransit"`
	DeliveryByTime        string `json:"DeliveryByTime,omitempty"`
}

type RatedPackage struct {
	TransportationCharges MonetaryValue    `json:"TransportationCharges"`
	BaseServiceCharge     MonetaryValue    `json:"BaseServiceCharge"`
	ServiceOptionsCharges MonetaryValue    `json:"ServiceOptionsCharges"`
	ItemizedCharges       []ItemizedCharge `json:"ItemizedCharges"`
	TotalCharges          MonetaryValue    `json:"TotalCharges"`
	Weight                string           `json:"Weight"`
	BillingWeight         Weight           `json:"BillingWeight"`
	SimpleRate            SimpleRateCode   `json:"SimpleRate"`
}

type ItemizedCharge struct {
	Code          string `json:"Code"`
	CurrencyCode  string `json:"CurrencyCode"`
	MonetaryValue string `json:"MonetaryValue"`
}

type SimpleRateCode struct {
	Code string `json:"Code"`
}