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
