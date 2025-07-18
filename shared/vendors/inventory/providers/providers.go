package providers

type (
	ProviderType string
)

const (
	ProviderNone       ProviderType = "none"
	ProviderSalesforce ProviderType = "salesforce"
	ProviderPrintful   ProviderType = "printful"
)
