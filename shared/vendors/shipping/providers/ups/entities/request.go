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
