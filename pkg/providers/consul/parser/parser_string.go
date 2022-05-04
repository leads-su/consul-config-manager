package parser

import "github.com/leads-su/logger"

// processString processes STRING and appends it to live data map
func (parser *Parser) processString(key, value string, delayed interface{}) {
	if parser.shouldDelay(delayed) {
		delayedPublisher, err := parser.newDelayedPublisher(value, delayed.(string))
		if err != nil {
			logger.Errorf("consul:parser:string", "failed to set delayed publisher for `%s` - %s", key, err.Error())
		} else {
			parser.setDelayedDataValue(key, delayedPublisher)
		}
	} else {
		parser.setDataValue(key, value)
	}
}
