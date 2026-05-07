package apperror

type AppError struct {
	Code            string `json:"code"`
	Message         string `json:"message"`
	TechnicalDetail string `json:"technicalDetail"`
	Suggestion      string `json:"suggestion"`
	Retryable       bool   `json:"retryable"`
}

func New(code, message, detail, suggestion string, retryable bool) *AppError {
	return &AppError{Code: code, Message: message, TechnicalDetail: detail, Suggestion: suggestion, Retryable: retryable}
}

func Wrap(code, message string, err error, suggestion string, retryable bool) *AppError {
	detail := ""
	if err != nil {
		detail = err.Error()
	}
	return New(code, message, detail, suggestion, retryable)
}

func (e *AppError) Error() string {
	if e == nil {
		return ""
	}
	return e.Code + ": " + e.Message
}
