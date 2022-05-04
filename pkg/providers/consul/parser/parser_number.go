package parser

import "github.com/leads-su/logger"

// processNumber processes NUMBER and appends it to live data map
func (parser *Parser) processNumber(key string, value interface{}, delayed interface{}) {
	if parser.shouldDelay(delayed) {
		delayedPublisher, err := parser.newDelayedPublisher(value, delayed.(string))
		if err != nil {
			logger.Errorf("consul:parser:number", "failed to set delayed publisher for `%s` - %s", key, err.Error())
		} else {
			parser.setDelayedDataValue(key, delayedPublisher)
		}
	} else {
		parser.setDataValue(key, value)
	}
}
