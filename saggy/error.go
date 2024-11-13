package saggy

type SaggyError struct {
	Message string
	Err     error
}

func (e *SaggyError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	} else {
		return e.Message
	}
}

func NewSaggyError(message string, err error) *SaggyError {
	return &SaggyError{Message: message, Err: err}
}
