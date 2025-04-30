package examples

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExamplesRandPrefix(t *testing.T) {
	withPrefix := examplesRandPrefix()
	kafkaName := withPrefix("kafka")

	// Starts with prefix, ends with "-kafka", has valid length
	assert.True(t, strings.HasPrefix(kafkaName, exampleTestsPrefix))
	assert.True(t, strings.HasSuffix(kafkaName, "-kafka"))
	assert.Len(t, exampleTestsPrefix+"0000000-kafka", len(kafkaName))

	// Equal seed for a new name
	mysqlName := withPrefix("mysql")
	assert.True(t, strings.HasPrefix(kafkaName, mysqlName[:len(mysqlName)-len("mysql")]))
}
