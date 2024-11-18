package api

// SubscriberGetSubscriberStatus representa a estrutura de resposta da API para obter o status de um assinante
type SubscriberGetSubscriberStatus struct {
	SubscriberUId string `json:"SubscriberUId"`
	AccessActive  string `json:"AccessActive"`
	CSPeriod      string `json:"CSPeriod"`
	AddInfo       struct {
		BillingDay       int64  `json:"BillingDay"`
		PaymentForm      string `json:"PaymentForm"`
		SubscriptionDate string `json:"SubscriptionDate"`
		BusinessType     string `json:"BusinessType"`
		SubscriptionType string `json:"SubscriptionType"`
	} `json:"AddInfo"`
}
