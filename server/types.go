package server

type PostVerifyRequest struct {
	Id            string `json:"id"`
	Domain        string `json:"domain"`
	ExpectedTXT   string `json:"expectedTXT"`
	ExpectedCNAME string `json:"expectedCNAME"`
}

type PostWebhookRequest struct {
	Id            string `json:"id"`
	Domain        string `json:"domain"`
	AcquiredTXT   string `json:"acquired_txt"`
	AcquiredCNAME string `json:"acquired_cname"`
}
