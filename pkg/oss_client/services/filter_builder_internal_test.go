package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLanguageLabel(t *testing.T) {
	assert.Equal(t, "Nederlands", languageLabel("nl"))
	assert.Equal(t, "Klingon", languageLabel("Klingon"))
}
