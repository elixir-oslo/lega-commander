package conf

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	teardown()
	os.Exit(code)
}

func setup() {
	_ = os.Setenv("CENTRAL_EGA_USERNAME", "1")
	_ = os.Setenv("CENTRAL_EGA_PASSWORD", "2")
	_ = os.Setenv("ELIXIR_AAI_TOKEN", "3")
}

func TestNewConfigurationSameInstance(t *testing.T) {
	configuration1 := NewConfiguration()
	configuration2 := NewConfiguration()
	println(configuration1)
	println(configuration2)
	if configuration1 != configuration2 {
		t.Error()
	}
}

func TestGetCentralEGAUsername(t *testing.T) {
	configuration := NewConfiguration()
	if configuration.GetCentralEGAUsername() != "1" {
		t.Error()
	}
}

func TestGetCentralEGAPassword(t *testing.T) {
	configuration := NewConfiguration()
	if configuration.GetCentralEGAPassword() != "2" {
		t.Error()
	}
}

func TestGetElixirAAIToken(t *testing.T) {
	configuration := NewConfiguration()
	if configuration.GetElixirAAIToken() != "3" {
		t.Error()
	}
}

func TestNewConfigurationDefaultInstanceURL(t *testing.T) {
	configuration := NewConfiguration()
	if configuration.GetLocalEGAInstanceURL() != defaultInstanceURL {
		t.Error()
	}
}

func TestNewConfigurationNonDefaultInstanceURL(t *testing.T) {
	_ = os.Setenv("LOCAL_EGA_INSTANCE_URL", "test/")
	configuration := NewConfiguration()
	if configuration.GetLocalEGAInstanceURL() != "test" {
		t.Error()
	}
}

func TestNewConfigurationDefaultChunkSize(t *testing.T) {
	configuration := NewConfiguration()
	if configuration.GetChunkSize() != defaultChunkSize {
		t.Error()
	}
}

func TestNewConfigurationNonDefaultChunkSize(t *testing.T) {
	_ = os.Setenv("LEGA_UPLOADER_CHUNK_SIZE", "100")
	configuration := NewConfiguration()
	if configuration.GetChunkSize() != 100 {
		t.Error()
	}
}

func TestNewConfigurationNonNumericChunkSize(t *testing.T) {
	_ = os.Setenv("LEGA_UPLOADER_CHUNK_SIZE", "test")
	configuration := NewConfiguration()
	if configuration.GetChunkSize() != defaultChunkSize {
		t.Error()
	}
}

func teardown() {
	_ = os.Unsetenv("CENTRAL_EGA_USERNAME")
	_ = os.Unsetenv("CENTRAL_EGA_PASSWORD")
	_ = os.Unsetenv("LOCAL_EGA_INSTANCE_URL")
	_ = os.Unsetenv("ELIXIR_AAI_TOKEN")
	_ = os.Unsetenv("LEGA_UPLOADER_CHUNK_SIZE")
}
