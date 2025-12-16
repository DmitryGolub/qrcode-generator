package http

type ErrorResponse struct {
	Success bool         `json:"success"`
	Error   ErrorDetails `json:"error"`
}

type ErrorDetails struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
