package snclient

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSNVersion(t *testing.T) {
	t.Parallel()
	snc := &Agent{}
	agent = snc
	res := snc.RunCheck("check_snclient_version", []string{})
	assert.Equalf(t, CheckExitOK, res.State, "state OK")
	assert.Regexpf(t, regexp.MustCompile("^SNClient"), res.Output, "output matches")
}