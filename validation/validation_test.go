package validation_test

import (
	"testing"

	"github.com/sillen102/simba/simbaTest/assert"
	"github.com/sillen102/simba/validation"
)

func TestValidateStruct_NilRequestReturnsNil(t *testing.T) {
	t.Parallel()

	assert.Nil(t, validation.ValidateStruct(nil))
}

func TestValidateStruct_ValidStructReturnsNil(t *testing.T) {
	t.Parallel()

	type validRequest struct {
		Name string `json:"name" validate:"required"`
		Age  int    `json:"age" validate:"gte=0"`
	}

	errors := validation.ValidateStruct(validRequest{
		Name: "Jane",
		Age:  42,
	})

	assert.Nil(t, errors)
}

func TestValidateStruct_UsesJsonTagForFieldName(t *testing.T) {
	t.Parallel()

	type request struct {
		FirstName string `json:"first_name" validate:"required"`
	}

	errors := validation.ValidateStruct(request{})

	assert.NotNil(t, errors)
	assert.Equal(t, 1, len(errors))
	assert.Equal(t, "first_name", errors[0].Field)
	assert.NotEqual(t, "", errors[0].Err)
}

func TestValidateStruct_FallsBackToStructFieldNameWithoutJsonTag(t *testing.T) {
	t.Parallel()

	type request struct {
		FirstName string `validate:"required"`
	}

	errors := validation.ValidateStruct(request{})

	assert.NotNil(t, errors)
	assert.Equal(t, 1, len(errors))
	assert.Equal(t, "FirstName", errors[0].Field)
	assert.NotEqual(t, "", errors[0].Err)
}

func TestValidateStruct_PointerInputUsesJsonTagFieldName(t *testing.T) {
	t.Parallel()

	type request struct {
		Email string `json:"email" validate:"required"`
	}

	req := &request{}
	errors := validation.ValidateStruct(req)

	assert.NotNil(t, errors)
	assert.Equal(t, 1, len(errors))
	assert.Equal(t, "email", errors[0].Field)
	assert.NotEqual(t, "", errors[0].Err)
}
