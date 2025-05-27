package entities

type XAVRequestWrapper struct {
	XAVRequest XAVRequest `json:"XAVRequest"`
}

type XAVRequest struct {
	AddressKeyFormat AddressKeyFormat `json:"AddressKeyFormat"`
}

type AddressKeyFormat struct {
	ConsigneeName       string   `json:"ConsigneeName,omitempty"`
	BuildingName        string   `json:"BuildingName,omitempty"`
	AddressLine         []string `json:"AddressLine"`
	Region              string   `json:"Region,omitempty"`
	PoliticalDivision2  string   `json:"PoliticalDivision2,omitempty"`
	PoliticalDivision1  string   `json:"PoliticalDivision1,omitempty"`
	PostcodePrimaryLow  string   `json:"PostcodePrimaryLow,omitempty"`
	PostcodeExtendedLow string   `json:"PostcodeExtendedLow,omitempty"`
	Urbanization        string   `json:"Urbanization,omitempty"`
	CountryCode         string   `json:"CountryCode,omitempty"`
}

type RateRequestWrapper struct {
	RateRequest RateRequest `json:"RateRequest"`
}

type RateRequest struct {
	Request  Request  `json:"Request"`
	Shipment Shipment `json:"Shipment"`
}

type Request struct {
	TransactionReference TransactionReference `json:"TransactionReference"`
}

type TransactionReference struct {
	CustomerContext string `json:"CustomerContext"`
}

type Shipment struct {
	Shipper     Party   `json:"Shipper"`
	ShipTo      Party   `json:"ShipTo"`
	NumOfPieces string  `json:"NumOfPieces"`
	Package     Package `json:"Package"`
}

type Party struct {
	Name          string  `json:"Name"`
	ShipperNumber string  `json:"ShipperNumber,omitempty"`
	Address       Address `json:"Address"`
}

type Address struct {
	AddressLine       []string `json:"AddressLine"`
	City              string   `json:"City"`
	StateProvinceCode string   `json:"StateProvinceCode"`
	PostalCode        string   `json:"PostalCode"`
	CountryCode       string   `json:"CountryCode"`
}

type Package struct {
	PackagingType CodeDescription `json:"PackagingType"`
	Dimensions    Dimensions      `json:"Dimensions"`
	PackageWeight PackageWeight   `json:"PackageWeight"`
}

type CodeDescription struct {
	Code        string `json:"Code"`
	Description string `json:"Description"`
}

type Dimensions struct {
	UnitOfMeasurement CodeDescription `json:"UnitOfMeasurement"`
	Length            string          `json:"Length"`
	Width             string          `json:"Width"`
	Height            string          `json:"Height"`
}

type PackageWeight struct {
	UnitOfMeasurement CodeDescription `json:"UnitOfMeasurement"`
	Weight            string          `json:"Weight"`
}
