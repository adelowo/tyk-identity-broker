package data_loader

import (
	log "github.com/Sirupsen/logrus"
	"github.com/TykTechnologies/tyk-identity-broker/configuration"
	"github.com/TykTechnologies/tyk-identity-broker/tap"
)

var dataLogger = log.WithField("prefix", "DATA LOADER")

// DataLoader is an interface that defines how data is loded from a source into a AuthRegisterBackend interface store
type DataLoader interface {
	Init(conf interface{}) error
	LoadIntoStore(tap.AuthRegisterBackend) error
	Flush(tap.AuthRegisterBackend) error
}

func CreateDataLoader(config configuration.Configuration, ProfileFilename *string) (DataLoader, error) {
	var dataLoader DataLoader
	var loaderConf interface{}

	//default storage
	storageType := configuration.FILE

	if config.Storage != nil {
		storageType = config.Storage.StorageType
	}

	switch storageType {
		case configuration.MONGO:
			dataLoader = &MongoLoader{}
			mongoConf := config.Storage.MongoConf
			dialInfo, err := MongoDialInfo(mongoConf.MongoURL, mongoConf.MongoUseSSL, mongoConf.MongoSSLInsecureSkipVerify)
			if err != nil {
				dataLogger.Error("Error getting mongo settings: " + err.Error())
				return nil, err
			}
			loaderConf = MongoLoaderConf{
				DialInfo: dialInfo,
			}
		default:
			//default: FILE
			dataLoader = &FileLoader{}
			//pDir := path.Join(config.ProfileDir, *ProfileFilename)
			loaderConf = FileLoaderConf{
				FileName:   *ProfileFilename,
				ProfileDir: config.ProfileDir,
			}
	}

	err := dataLoader.Init(loaderConf)
	return dataLoader, err
}