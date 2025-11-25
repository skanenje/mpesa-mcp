package mpesa

// TokenResponse represents the OAuth token response from Daraja API
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   string `json:"expires_in"`
}

// STKPushRequest represents the request payload for STK Push
type STKPushRequest struct {
	BusinessShortCode string `json:"BusinessShortCode"`
	Password          string `json:"Password"`
	Timestamp         string `json:"Timestamp"`
	TransactionType   string `json:"TransactionType"`
	Amount            int    `json:"Amount"`
	PartyA            string `json:"PartyA"`
	PartyB            string `json:"PartyB"`
	PhoneNumber       string `json:"PhoneNumber"`
	CallBackURL       string `json:"CallBackURL"`
	AccountReference  string `json:"AccountReference"`
	TransactionDesc   string `json:"TransactionDesc"`
}

// STKPushResponse represents the response from STK Push API
type STKPushResponse struct {
	MerchantRequestID   string `json:"MerchantRequestID"`
	CheckoutRequestID   string `json:"CheckoutRequestID"`
	ResponseCode        string `json:"ResponseCode"`
	ResponseDescription string `json:"ResponseDescription"`
	CustomerMessage     string `json:"CustomerMessage"`
	ErrorCode           string `json:"errorCode,omitempty"`
	ErrorMessage        string `json:"errorMessage,omitempty"`
}

// QRCodeRequest represents the request for QR code generation
type QRCodeRequest struct {
	MerchantName string `json:"MerchantName"`
	RefNo        string `json:"RefNo"`
	Amount       int    `json:"Amount"`
	TrxCode      string `json:"TrxCode"`
	CPI          string `json:"CPI"`
	Size         string `json:"Size"`
}

// QRCodeResponse represents the response from QR code generation
type QRCodeResponse struct {
	ResponseCode        string `json:"ResponseCode"`
	RequestID           string `json:"RequestID"`
	ResponseDescription string `json:"ResponseDescription"`
	QRCode              string `json:"QRCode"`
	ErrorCode           string `json:"errorCode,omitempty"`
	ErrorMessage        string `json:"errorMessage,omitempty"`
}
