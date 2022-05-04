package parser

import "time"

type DelayedPublishing struct {
	Value     interface{}
	PublishOn time.Time
}

// newDelayedPublisher creates new instance of delayed publisher
func (parser *Parser) newDelayedPublisher(value interface{}, publishOn string) (*DelayedPublishing, error) {
	publishOnTime, err := time.Parse(time.RFC3339, publishOn)
	if err != nil {
		return nil, err
	}
	return &DelayedPublishing{
		Value:     value,
		PublishOn: publishOnTime,
	}, nil
}

// shouldPublish checks whether value should be published
func (dp *DelayedPublishing) shouldPublish() bool {
	localTime := time.Now().UTC().Unix()
	publishTime := dp.PublishOn.UTC().Unix()
	return localTime >= publishTime
}
