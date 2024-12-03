package simba_test

import (
	"testing"

	"github.com/sillen102/simba"
	"gotest.tools/v3/assert"
)

func TestSettingOptions(t *testing.T) {
	t.Parallel()

	t.Run("default options", func(t *testing.T) {
		router := simba.NewRouter()
		assert.Equal(t, router.GetOptions().RequestDisallowUnknownFields, true)
	})

	t.Run("set disallow unknown fields", func(t *testing.T) {
		options := simba.RouterOptions{
			RequestDisallowUnknownFields: false,
		}
		router := simba.NewRouter(options)

		assert.Equal(t, router.GetOptions().RequestDisallowUnknownFields, options.RequestDisallowUnknownFields)
	})
}
