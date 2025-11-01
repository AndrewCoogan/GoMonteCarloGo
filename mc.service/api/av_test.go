package api

import (
	"os"
	"testing"

	"github.com/joho/godotenv"
)

func Test_GetAlphaVantageApiKey(t *testing.T) {
	err := godotenv.Load("testenv");
	if err != nil {
		t.Errorf("error loading environment: %s", err)
	}

	keyName := "ALPHAVANTAGE_API_KEY"
	actual := os.Getenv(keyName)
	if actual == "" {
		t.Errorf("error finding key %s in .env", keyName)
	}

	expected := "av-test-api-key"
	if actual != expected {
		t.Errorf("error validating key. expected %s, got %s", expected, actual)
	}
}