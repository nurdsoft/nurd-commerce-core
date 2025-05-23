package entities

type ShippingRateRequest struct {
	CarrierIds        []string         `json:"carrier_ids,omitempty"`
	FromCountryCode   string           `json:"from_country_code,omitempty"`
	FromPostalCode    string           `json:"from_postal_code,omitempty"`
	FromCityLocality  string           `json:"from_city_locality,omitempty"`
	FromStateProvince string           `json:"from_state_province,omitempty"`
	ToCountryCode     string           `json:"to_country_code,omitempty"`
	ToPostalCode      string           `json:"to_postal_code,omitempty"`
	ToCityLocality    string           `json:"to_city_locality,omitempty"`
	ToStateProvince   string           `json:"to_state_province,omitempty"`
	Weight            Weight           `json:"weight,omitempty"`
	Dimensions        ObjectDimensions `json:"dimensions,omitempty"`
}

type Weight struct {
	Value float64 `json:"value,omitempty"`
	Unit  string  `json:"unit,omitempty"`
}

type ObjectDimensions struct {
	Length float64 `json:"length,omitempty"`
	Width  float64 `json:"width,omitempty"`
	Height float64 `json:"height,omitempty"`
	Unit   string  `json:"unit,omitempty"`
}
