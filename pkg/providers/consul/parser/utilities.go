package parser

import (
	"fmt"
	"strings"
)

// formatKey formats key and makes it a valid env variable
func (parser *Parser) formatKey(key string) string {
	return "CONSUL_" + strings.ReplaceAll(strings.ReplaceAll(strings.ToUpper(key), "/", "_"), "-", "_")
}

// setReferenceValue adds data to reference map
func (parser *Parser) setReferenceValue(key, toKey string) {
	parser.Lock()
	defer parser.Unlock()
	parser.referenceMap[key] = toKey
}

// removeReferenceValue removes data from reference map
func (parser *Parser) removeReferenceValue(key string) {
	parser.Lock()
	defer parser.Unlock()
	referenceMap := make(map[string]string)
	for index, value := range parser.referenceMap {
		if index != key {
			referenceMap[index] = value
		}
	}
	parser.referenceMap = referenceMap
}

// getDataValue retrieves data from live data map
func (parser *Parser) getDataValue(key, reference string) (interface{}, error) {
	parser.Lock()
	defer parser.Unlock()
	if value, ok := parser.liveData[reference]; ok {
		return value, nil
	}
	return nil, fmt.Errorf("unable to find live data value for key `%s` which references `%s`", key, reference)
}

// setDataValue adds data to live data map
func (parser *Parser) setDataValue(key string, value interface{}) {
	parser.Lock()
	defer parser.Unlock()
	parser.liveData[key] = value
}

// removeDataValue removes value from live data map
func (parser *Parser) removeDataValue(key string) {
	parser.Lock()
	defer parser.Unlock()
	liveData := make(map[string]interface{})
	for index, value := range parser.liveData {
		if index != key {
			liveData[index] = value
		}
	}
	parser.liveData = liveData
}

// setDelayedDataValue adds value to delayed data map
func (parser *Parser) setDelayedDataValue(key string, value *DelayedPublishing) {
	parser.Lock()
	defer parser.Unlock()
	parser.delayedData[key] = value
}

// removeDelayedDataValue removes value from delayed data map
func (parser *Parser) removeDelayedDataValue(key string) {
	parser.Lock()
	defer parser.Unlock()
	delayedData := make(map[string]*DelayedPublishing)
	for index, value := range parser.delayedData {
		if index != key {
			delayedData[key] = value
		}
	}
	parser.delayedData = delayedData
}
