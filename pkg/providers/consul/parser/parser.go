package parser

import (
	"fmt"
	"sync"

	"github.com/hashicorp/consul/api"
	"github.com/leads-su/logger"
)

type Parser struct {
	sync.RWMutex
	referenceMap     map[string]string
	liveData         map[string]interface{}
	delayedData      map[string]*DelayedPublishing
	referenceStorage *ReferenceStorage
}

func NewParser() *Parser {
	return &Parser{
		referenceMap:     make(map[string]string),
		liveData:         make(map[string]interface{}),
		delayedData:      make(map[string]*DelayedPublishing),
		referenceStorage: NewReferenceStorage(),
	}
}

// ProcessReceivedData process data received from Consul
func (parser *Parser) ProcessReceivedData(pairs api.KVPairs) {
	for _, entry := range pairs {
		if entry.Value != nil {
			key := parser.formatKey(entry.Key)
			parser.referenceStorage.Set(entry.Key, key)
			value := parser.processConsulValue(entry.Key, entry.Value)

			if value.Type == "reference" {
				targetKey := parser.formatKey(fmt.Sprintf("%v", value.Value))
				parser.setReferenceValue(key, targetKey)
			} else {
				parser.processValue(key, value)
			}
		}
	}
}

// GenerateConfiguration tries to process delayed and reference data and generates configuration
func (parser *Parser) GenerateConfiguration() map[string]interface{} {
	for key, delayedPublisher := range parser.delayedData {
		if delayedPublisher.shouldPublish() {
			parser.setDataValue(key, delayedPublisher.Value)
			parser.removeDelayedDataValue(key)
		} else {
			parser.setDataValue(key, "consul_delayed_publishing")
		}
	}

	// This cycle is needed to be able to resolve the target value for the nested references
	// Due to the fact, that one reference can point to another, we need to get to the target
	// value which they are referencing, this code is covering that use case
	for key := range parser.referenceMap {
		initialValue := key
		targetValue := key

		_, existsInMap := parser.referenceMap[targetValue]
		for existsInMap {
			targetValue = parser.referenceMap[targetValue]
			_, existsInMap = parser.referenceMap[targetValue]
		}

		val, err := parser.getDataValue(initialValue, targetValue)
		if err != nil {
			logger.Errorf("consul:parser", "failed to retrieve reference value - %s", err)
			continue
		}
		parser.setDataValue(initialValue, val)
		parser.removeReferenceValue(initialValue)
	}

	return parser.liveData
}

// GetReferenceStorage returns instance of reference storage
func (parser *Parser) GetReferenceStorage() *ReferenceStorage {
	return parser.referenceStorage
}

// shouldDelay simply checks if delayed parameter is not equal to null (not null = should be delayed)
func (parser *Parser) shouldDelay(delayed interface{}) bool {
	return delayed != nil
}
