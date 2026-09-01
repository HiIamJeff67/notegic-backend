package kafka

type ErrorClassification string

const (
	ErrorClassification_Transient          ErrorClassification = "Transient"
	ErrorClassification_PoisonMessage      ErrorClassification = "PoisonMessage"
	ErrorClassification_SchemaIncompatible ErrorClassification = "SchemaIncompatible"
)

type ConsumerError struct {
	Classification ErrorClassification
	Origin         error
}

func (e *ConsumerError) Error() string {
	if e == nil || e.Origin == nil {
		return "Kafka consumer error"
	}

	return e.Origin.Error()
}

func (e *ConsumerError) Unwrap() error {
	if e == nil {
		return nil
	}

	return e.Origin
}
