package parser

import (
	"encoding/json"

	"github.com/leads-su/logger"
)

type ConsulValue struct {
	Type    string      `json:"type"`
	Delayed interface{} `json:"delayed"`
	Value   interface{} `json:"value"`
}

// processConsulValue decodes value received from Consul into struct
func (parser *Parser) processConsulValue(key string, value []byte) *ConsulValue {
	var processedValue *ConsulValue

	err := json.Unmarshal(value, &processedValue)
	if err != nil {
		logger.Errorf("consul:parser", "failed to decode value for `%s` key - %s", key, err.Error())
		return nil
	}

	return processedValue
}

// processValue processes decoded value received from Consul
func (parser *Parser) processValue(key string, value *ConsulValue) {
	switch value.Type {
	case "array":
		parser.processArray(key, value.Value.([]interface{}), value.Delayed)
		break
	case "number":
		parser.processNumber(key, value.Value, value.Delayed)
		break
	case "string":
		parser.processString(key, value.Value.(string), value.Delayed)
		break
	default:
		logger.Errorf("consul:parser", "Unknown value type - %s (%T)", value.Type, value.Type)
	}
}
