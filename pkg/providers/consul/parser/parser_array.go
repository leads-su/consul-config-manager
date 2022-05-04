package parser

import (
	"fmt"
	"strings"

	"github.com/leads-su/logger"
)

// processArray processes ARRAY and appends it to live data map
func (parser *Parser) processArray(key string, values []interface{}, delayed interface{}) {
	var newValue []string
	for _, value := range values {
		newValue = append(newValue, fmt.Sprintf("%v", value))
	}

	if parser.shouldDelay(delayed) {
		delayedPublisher, err := parser.newDelayedPublisher(values, delayed.(string))
		if err != nil {
			logger.Errorf("consul:parser:array", "failed to set delayed publisher for `%s` - %s", key, err.Error())
		} else {
			parser.setDelayedDataValue(key, delayedPublisher)
		}
	} else {
		parser.setDataValue(key, strings.Join(newValue, "\n"))
	}
}
