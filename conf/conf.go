// Package conf contains structure and methods to create/load configuration for the application.
package conf

import (
	"github.com/logrusorgru/aurora"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
)

const defaultInstanceURL = "https://ega.elixir.no"
const defaultChunkSize = 50

var once sync.Once
var instance *defaultConfiguration

// Configuration interface is a holder for application settings.
type Configuration interface {
	GetCentralEGAUsername() string
	GetCentralEGAPassword() string
	GetLocalEGAInstanceURL() string
	GetElixirAAIToken() string
	GetChunkSize() int
}

// defaultConfiguration structure is a default implementation of the Configuration interface.
type defaultConfiguration struct {
}

func (dc defaultConfiguration) GetCentralEGAUsername() string {
	centralEGAUsername := os.Getenv("CENTRAL_EGA_USERNAME")
	if centralEGAUsername == "" {
		log.Fatal(aurora.Red("CENTRAL_EGA_USERNAME environment variable is not set"))
	}
	return centralEGAUsername
}

func (dc defaultConfiguration) GetCentralEGAPassword() string {
	centralEGAPassword := os.Getenv("CENTRAL_EGA_PASSWORD")
	if centralEGAPassword == "" {
		log.Fatal(aurora.Red("CENTRAL_EGA_PASSWORD environment variable is not set"))
	}
	return centralEGAPassword
}

func (dc defaultConfiguration) GetLocalEGAInstanceURL() string {
	localEGAInstanceURL := os.Getenv("LOCAL_EGA_INSTANCE_URL")
	if localEGAInstanceURL == "" {
		localEGAInstanceURL = defaultInstanceURL
	}
	if strings.HasSuffix(localEGAInstanceURL, "/") {
		return localEGAInstanceURL[:len(localEGAInstanceURL)-1]
	}
	return localEGAInstanceURL
}

func (dc defaultConfiguration) GetElixirAAIToken() string {
	elixirAAIToken := os.Getenv("ELIXIR_AAI_TOKEN")
	if elixirAAIToken == "" {
		log.Fatal(aurora.Red("ELIXIR_AAI_TOKEN environment variable is not set"))
	}
	return elixirAAIToken
}

func (dc defaultConfiguration) GetChunkSize() int {
	chunkSize := os.Getenv("LEGA_UPLOADER_CHUNK_SIZE")
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
