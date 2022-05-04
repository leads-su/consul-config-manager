package storage

import (
	"bufio"
	"bytes"
	"fmt"
	cfg "github.com/leads-su/consul-config-manager/pkg/config"
	"github.com/leads-su/consul-config-manager/pkg/config/notifier"
	p "github.com/leads-su/consul-config-manager/pkg/providers/consul/parser"
	"github.com/leads-su/logger"
	notifierPackage "github.com/leads-su/notifier"
	s "github.com/leads-su/storage"
	"os"
	"strconv"
	"strings"
	"sync"
	"text/template"
)

type ConsulStorage struct {
	sync.RWMutex
	// config is an instance of application configuration
	config *cfg.Config

	// storage is an instance of storage
	storage *s.Storage

	// parser is an instance of parser
	parser *p.Parser
}

// NewStorage create new Consul storage instance
func NewStorage(config *cfg.Config, parser *p.Parser) *ConsulStorage {
	return &ConsulStorage{
		config: config,
		storage: s.NewStorage(s.Options{
			WorkingDirectory: config.Consul.WriteTo,
		}),
		parser: parser,
	}
}

type ConfigContent = map[string]interface{}
type Configs = map[string]ConfigContent

// ProcessChanges processes changes retrieved from Consul
func (cs *ConsulStorage) ProcessChanges(changes map[string]interface{}) {
	cs.Lock()
	defer cs.Unlock()
	configs := make(Configs)
	for k, v := range changes {
		configPath := cs.generateConfigurationFilePath(k)
		if _, ok := configs[configPath]; !ok {
			configs[configPath] = make(ConfigContent)
		}
		configs[configPath][k] = v
	}
	for path, variables := range configs {
		cs.writeToFile(path, variables)
	}
}

// writeToFile writes data to file
func (cs *ConsulStorage) writeToFile(path string, variables ConfigContent) {
	var fileLines []string

	for key, value := range variables {
		switch value.(type) {
		case bool:
			fileLines = append(fileLines, key+"="+fmt.Sprintf("%t", value))
		case float32, float64:
			fileLines = append(fileLines, key+"="+strconv.FormatFloat(value.(float64), 'f', -1, 64))
		case int, int8, int16, int32, int64:
			fileLines = append(fileLines, key+"="+fmt.Sprintf("%d", value))
		case string:
			formattedValue := fmt.Sprintf("%s", value)
			fileLines = append(fileLines, key+"="+strconv.Quote(formattedValue))
		}
	}

	tempFileHash, err := cs.writeToTempFile(path, fileLines)
	if err == nil {
		backupFilePath := fmt.Sprintf("%s.bak", path)
		if cs.storage.Exists(cs.storage.AbsolutePath(path)) {
			err = cs.storage.MoveFile(cs.storage.AbsolutePath(path), cs.storage.AbsolutePath(backupFilePath))
			if err != nil {
				errMsg := fmt.Sprintf("failed to create file backup (%s) - %s", path, err.Error())
				logger.Errorf("consul:storage", errMsg)
				cs.sendErrorNotification(errMsg)
			}
		}
		err = cs.storage.CreateDirectory(cs.storage.AbsolutePath(path))
		if err != nil {
			logger.Errorf("consul:storage", "failed to create data path")
			cs.sendErrorNotification("failed to create data path")
		} else {
			err = cs.storage.MoveFile(cs.storage.AbsoluteTempPath(path), cs.storage.AbsolutePath(path))
			if err != nil {
				errMsg := fmt.Sprintf("failed to move `%s` from temporary folder to permanent location - %s", path, err.Error())
				logger.Error("consul:storage", errMsg)
				cs.sendErrorNotification(errMsg)

				err = cs.storage.MoveFile(cs.storage.AbsolutePath(backupFilePath), cs.storage.AbsolutePath(path))
				if err != nil {
					errMsg := fmt.Sprintf("failed to restore file backup (%s) - %s", path, err.Error())
					logger.Error("consul:storage", errMsg)
					cs.sendErrorNotification(errMsg)
				}
			} else {
				finalFileHash, err := cs.storage.ComputeFileHash(cs.storage.AbsolutePath(path))
				if err != nil {
					errMsg := fmt.Sprintf("failed to compute final file hash (%s) - %s", path, err.Error())
					logger.Error("consul:storage", errMsg)
					cs.sendErrorNotification(errMsg)
				} else {
					if tempFileHash != finalFileHash {
						err = cs.storage.MoveFile(cs.storage.AbsolutePath(backupFilePath), cs.storage.AbsolutePath(path))
						if err != nil {
							errMsg := fmt.Sprintf("failed to restore file backup (%s) - %s", path, err.Error())
							logger.Error("consul:storage", errMsg)
							cs.sendErrorNotification(errMsg)
						}
					}
				}
			}
		}
	}
}

// sendErrorNotification sends error notification
func (cs *ConsulStorage) sendErrorNotification(message string) {
	tpl, err := template.New("").Parse(`
*Message:*     {{ .Message }}
*Application:* {{ .Application }}
*Version:*     {{ .Version }}-{{ .ShaHash }}
*Server:*      {{ .Address }}
`)
	if err == nil {
		templateValues := struct {
			Message     string
			Application string
			Version     string
			ShaHash     string
			Address     string
		}{
			Message:     message,
			Application: "Consul Config Manager",
			Version:     cs.config.Application.Version,
			ShaHash:     cs.config.Application.CommitSha,
			Address:     cs.config.Agent.Address(),
		}

		var templateBuffer bytes.Buffer
		err = tpl.Execute(&templateBuffer, templateValues)
		if err == nil {
			cs.config.Notifier.DeliverNotification(notifier.TELEGRAM, notifierPackage.NewNotification(notifierPackage.NotificationOptions{
				Type:    notifierPackage.Error,
				Title:   fmt.Sprintf("%s - consul:storage error", cs.config.Agent.Network.Hostname()),
				Message: templateBuffer.String(),
			}))
		}
	}
}

// writeToTempFile writes data to temporary file
func (cs *ConsulStorage) writeToTempFile(path string, fileLines []string) (string, error) {
	absolutePath := cs.storage.AbsoluteTempPath(path)
	err := cs.storage.CreateFile(absolutePath)
	if err != nil {
		return "", err
	}
	file, err := os.OpenFile(absolutePath, os.O_RDWR|os.O_TRUNC, 0755)
	if err != nil {
		logger.Errorf("consul:storage", "failed to write configuration to file - %s", err)
		return "", err
	}

	dataWriter := bufio.NewWriter(file)

	for _, line := range fileLines {
		_, err := dataWriter.WriteString(line + "\n")
		if err != nil {
			logger.Warnf("consul:storage", "failed to write line to file - %s", err)
		}
	}

	if err = dataWriter.Flush(); err != nil {
		logger.Warnf("consul:storage", "failed to flush data writer - %s", err)
		return "", err
	}

	if err = file.Close(); err != nil {
		logger.Fatalf("consul:storage", "failed to close opened file - %s", err)
		return "", err
	}

	hash, err := cs.storage.ComputeFileHash(absolutePath)
	if err != nil {
		logger.Errorf("consul:storage", "failed to compute hash for a file - %s", err.Error())
		return "", err
	}
	return hash, nil
}

// generateConfigurationFilePath generates OS independent path to configuration file
func (cs *ConsulStorage) generateConfigurationFilePath(key string) string {
	path, err := cs.parser.GetReferenceStorage().Get(key)
	if err != nil {
		logger.Fatalf("provider:consul:storage", "failed to retrieve key reference to path")
	}
	stringParts := strings.SplitN(path, "/", -1)
	return fmt.Sprintf("%s%s%s.env", strings.ToLower(stringParts[0]), string(os.PathSeparator), strings.ToLower(stringParts[1]))
}
