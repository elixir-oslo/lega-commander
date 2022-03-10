// Package conf contains structure and methods to create/load configuration for the application.
package conf

import (
	"log"
	"os"
	"strconv"
	"strings"
	"sync"

	aurora "github.com/logrusorgru/aurora/v3"
)

const defaultInstanceURL = "https://ega.elixir.no"
const defaultTSDFileAPIVersion = "v1"
const defaultTSDService = "ega"
const defaultTSDProject = "p969"
const defaultTSDfileAPIbaseURL = "https://api.tsd.usit.no"
const defaultChunkSize = 50

var once sync.Once
var instance *defaultConfiguration

// Configuration interface is a holder for application settings.
type Configuration interface {
	ConcatenateURLPartsToString(array []string) string
	GetTSDbaseURL() string
	GetTSDAPIVersion() string
	GetTSDProjectName() string
	GetTSDservice() string
	GetTSDURL() string
	GetCentralEGAUsername() string
	GetCentralEGAPassword() string
	GetLocalEGAInstanceURL() string
	GetElixirAAIToken() string
	GetChunkSize() int
}

func (defaultConfiguration) ConcatenateURLPartsToString(array []string) string {
	str := strings.Join(array, "/")
	return str
}

// defaultConfiguration structure is a default implementation of the Configuration interface.
type defaultConfiguration struct {
}

func (defaultConfiguration) GetCentralEGAUsername() string {
	centralEGAUsername := os.Getenv("CENTRAL_EGA_USERNAME")
	if centralEGAUsername == "" {
		log.Fatal(aurora.Red("CENTRAL_EGA_USERNAME environment variable is not set"))
	}
	return centralEGAUsername
}

func (defaultConfiguration) GetCentralEGAPassword() string {
	centralEGAPassword := os.Getenv("CENTRAL_EGA_PASSWORD")
	if centralEGAPassword == "" {
		log.Fatal(aurora.Red("CENTRAL_EGA_PASSWORD environment variable is not set"))
	}
	return centralEGAPassword
}

func (defaultConfiguration) GetLocalEGAInstanceURL() string {
	localEGAInstanceURL := os.Getenv("LOCAL_EGA_INSTANCE_URL")
	if localEGAInstanceURL == "" {
		localEGAInstanceURL = defaultInstanceURL
	}
	if strings.HasSuffix(localEGAInstanceURL, "/") {
		return localEGAInstanceURL[:len(localEGAInstanceURL)-1]
	}
	return localEGAInstanceURL
}

func (defaultConfiguration) GetElixirAAIToken() string {
	elixirAAIToken := os.Getenv("ELIXIR_AAI_TOKEN")
	if elixirAAIToken == "" {
		log.Fatal(aurora.Red("ELIXIR_AAI_TOKEN environment variable is not set"))
	}
	return elixirAAIToken
}

func (dc defaultConfiguration) GetTSDURL() string {
	return dc.ConcatenateURLPartsToString(
		[]string{
			dc.GetTSDbaseURL(), dc.GetTSDAPIVersion(), dc.GetTSDProjectName(), dc.GetTSDservice()},
	)
}

func (defaultConfiguration) GetTSDbaseURL() string {
	TSDbaseURL := os.Getenv("TSD_BASE_URL")
	if TSDbaseURL == "" {
		TSDbaseURL = defaultTSDfileAPIbaseURL
	}
	if strings.HasSuffix(TSDbaseURL, "/") {
		return TSDbaseURL[:len(TSDbaseURL)-1]
	}
	return TSDbaseURL
}

func (defaultConfiguration) GetTSDAPIVersion() string {
	return defaultTSDFileAPIVersion
}

func (defaultConfiguration) GetTSDProjectName() string {
	tsdProject := os.Getenv("TSD_PROJ_NAME")
	if tsdProject == "" {
		tsdProject = defaultTSDProject
	}
	return tsdProject
}

func (defaultConfiguration) GetTSDservice() string {
	return defaultTSDService
}

func (defaultConfiguration) GetChunkSize() int {
	chunkSize := os.Getenv("LEGA_COMMANDER_CHUNK_SIZE")
	if chunkSize == "" {
		return defaultChunkSize
	}
	numericChunkSize, err := strconv.Atoi(chunkSize)
	if err != nil {
		return defaultChunkSize
	}
	return numericChunkSize
}

// NewConfiguration constructs Configuration, accepting LocalEGA URL instance and possibly chunk size.
func NewConfiguration() Configuration {
	once.Do(func() {
		instance = &defaultConfiguration{}
	})
	return instance
}
