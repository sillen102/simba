package simba_test

import (
	"testing"

	"github.com/sillen102/simba"
	"github.com/sillen102/simba/simbaTest/assert"
)

// Comparisons

func TestValidateStruct_Gte_Int(t *testing.T) {
	t.Parallel()

	type TestStructGte struct {
		Age int `json:"age" validate:"gte=0"`
	}

	// Given
	testStruct := TestStructGte{
		Age: -1,
	}

	// When
	errors := simba.ValidateStruct(testStruct)

	// Then
	assert.NotNil(t, errors)
	assert.Equal(t, 1, len(errors))
	assert.Equal(t, "age", errors[0].Field)
	assert.Equal(t, "age must be greater than or equal to 0", errors[0].Err)

	// Given
	testStruct = TestStructGte{
		Age: 0,
	}

	// When
	errors = simba.ValidateStruct(testStruct)

	// Then
	assert.Nil(t, errors)
}

func TestValidateStruct_Gte_Float(t *testing.T) {
	t.Parallel()

	type TestStructGte struct {
		Age float64 `json:"age" validate:"gte=0.0"`
	}

	// Given
	testStruct := TestStructGte{
		Age: -1,
	}

	// When
	errors := simba.ValidateStruct(testStruct)

	// Then
	assert.NotNil(t, errors)
	assert.Equal(t, 1, len(errors))
	assert.Equal(t, "age", errors[0].Field)
	assert.Equal(t, "age must be greater than or equal to 0.0", errors[0].Err)

	// Given
	testStruct = TestStructGte{
		Age: 0,
	}

	// When
	errors = simba.ValidateStruct(testStruct)

	// Then
	assert.Nil(t, errors)
}

func TestValidateStruct_Lte_Int(t *testing.T) {
	t.Parallel()

	type TestStructLte struct {
		Age int `json:"age" validate:"lte=0"`
	}

	// Given
	testStruct := TestStructLte{
		Age: 1,
	}

	// When
	errors := simba.ValidateStruct(testStruct)

	// Then
	assert.NotNil(t, errors)
	assert.Equal(t, 1, len(errors))
	assert.Equal(t, "age", errors[0].Field)
	assert.Equal(t, "age must be less than or equal to 0", errors[0].Err)

	// Given
	testStruct = TestStructLte{
		Age: 0,
	}

	// When
	errors = simba.ValidateStruct(testStruct)

	// Then
	assert.Nil(t, errors)
}

func TestValidateStruct_Lte_Float(t *testing.T) {
	t.Parallel()

	type TestStructLte struct {
		Age float64 `json:"age" validate:"lte=0.0"`
	}

	// Given
	testStruct := TestStructLte{
		Age: 1.0,
	}

	// When
	errors := simba.ValidateStruct(testStruct)

	// Then
	assert.NotNil(t, errors)
	assert.Equal(t, 1, len(errors))
	assert.Equal(t, "age", errors[0].Field)
	assert.Equal(t, "age must be less than or equal to 0.0", errors[0].Err)

	// Given
	testStruct = TestStructLte{
		Age: 0.0,
	}

	// When
	errors = simba.ValidateStruct(testStruct)

	// Then
	assert.Nil(t, errors)
}

func TestValidateStruct_Gt_Int(t *testing.T) {
	t.Parallel()

	type TestStructGt struct {
		Age int `json:"age" validate:"gt=0"`
	}

	// Given
	testStruct := TestStructGt{
		Age: 0,
	}

	// When
	errors := simba.ValidateStruct(testStruct)

	// Then
	assert.NotNil(t, errors)
	assert.Equal(t, 1, len(errors))
	assert.Equal(t, "age", errors[0].Field)
	assert.Equal(t, "age must be greater than 0", errors[0].Err)

	// Given
	testStruct = TestStructGt{
		Age: 1,
	}

	// When
	errors = simba.ValidateStruct(testStruct)

	// Then
	assert.Nil(t, errors)
}

func TestValidateStruct_Gt_Float(t *testing.T) {
	t.Parallel()

	type TestStructGt struct {
		Age float64 `json:"age" validate:"gt=0.0"`
	}

	// Given
	testStruct := TestStructGt{
		Age: 0.0,
	}

	// When
	errors := simba.ValidateStruct(testStruct)

	// Then
	assert.NotNil(t, errors)
	assert.Equal(t, 1, len(errors))
	assert.Equal(t, "age", errors[0].Field)
	assert.Equal(t, "age must be greater than 0.0", errors[0].Err)

	// Given
	testStruct = TestStructGt{
		Age: 1.0,
	}

	// When
	errors = simba.ValidateStruct(testStruct)

	// Then
	assert.Nil(t, errors)
}

func TestValidateStruct_Lt_Int(t *testing.T) {
	t.Parallel()

	type TestStructLt struct {
		Age int `json:"age" validate:"lt=0"`
	}

	// Given
	testStruct := TestStructLt{
		Age: 0,
	}

	// When
	errors := simba.ValidateStruct(testStruct)

	// Then
	assert.NotNil(t, errors)
	assert.Equal(t, 1, len(errors))
	assert.Equal(t, "age", errors[0].Field)
	assert.Equal(t, "age must be less than 0", errors[0].Err)

	// Given
	testStruct = TestStructLt{
		Age: -1,
	}

	// When
	errors = simba.ValidateStruct(testStruct)

	// Then
	assert.Nil(t, errors)
}

func TestValidateStruct_Lt_Float(t *testing.T) {
	t.Parallel()

	type TestStructLt struct {
		Age float64 `json:"age" validate:"lt=0.0"`
	}

	// Given
	testStruct := TestStructLt{
		Age: 0.0,
	}

	// When
	errors := simba.ValidateStruct(testStruct)

	// Then
	assert.NotNil(t, errors)
	assert.Equal(t, 1, len(errors))
	assert.Equal(t, "age", errors[0].Field)
	assert.Equal(t, "age must be less than 0.0", errors[0].Err)

	// Given
	testStruct = TestStructLt{
		Age: -1.0,
	}

	// When
	errors = simba.ValidateStruct(testStruct)

	// Then
	assert.Nil(t, errors)
}

// Strings

func TestValidateStruct_Alpha(t *testing.T) {
	t.Parallel()

	type TestStructAlpha struct {
		Name string `json:"name" validate:"alpha"`
	}

	// Given
	testStruct := TestStructAlpha{
		Name: "John123",
	}

	// When
	errors := simba.ValidateStruct(testStruct)

	// Then
	assert.NotNil(t, errors)
	assert.Equal(t, 1, len(errors))
	assert.Equal(t, "name", errors[0].Field)
	assert.Equal(t, "name must contain only letters", errors[0].Err)

	// Given
	testStruct = TestStructAlpha{
		Name: "Jane",
	}

	// When
	errors = simba.ValidateStruct(testStruct)

	// Then
	assert.Nil(t, errors)
}

func TestValidateStruct_Alphanum(t *testing.T) {
	t.Parallel()

	type TestStructAlphanum struct {
		Name string `json:"name" validate:"alphanum"`
	}

	// Given
	testStruct := TestStructAlphanum{
		Name: "JohnDoé123",
	}

	// When
	errors := simba.ValidateStruct(testStruct)

	// Then
	assert.NotNil(t, errors)
	assert.Equal(t, 1, len(errors))
	assert.Equal(t, "name", errors[0].Field)
	assert.Equal(t, "name must contain only letters and numbers", errors[0].Err)

	// Given
	testStruct = TestStructAlphanum{
		Name: "JaneDoe123",
	}

	// When
	errors = simba.ValidateStruct(testStruct)

	// Then
	assert.Nil(t, errors)
}

func TestValidateStruct_Alphanumunicode(t *testing.T) {
	t.Parallel()

	type TestStructAlphanumunicode struct {
		Name string `json:"name" validate:"alphanumunicode"`
	}

	// Given
	testStruct := TestStructAlphanumunicode{
		Name: "JohnDoé123%",
	}

	// When
	errors := simba.ValidateStruct(testStruct)

	// Then
	assert.NotNil(t, errors)
	assert.Equal(t, 1, len(errors))
	assert.Equal(t, "name", errors[0].Field)
	assert.Equal(t, "name must contain only letters and numbers that are part of unicode", errors[0].Err)

	// Given
	testStruct = TestStructAlphanumunicode{
		Name: "JaneDoé123",
	}

	// When
	errors = simba.ValidateStruct(testStruct)

	// Then
	assert.Nil(t, errors)
}

func TestValidateStruct_Alphaunicode(t *testing.T) {
	t.Parallel()

	type TestStructAlphaunicode struct {
		Name string `json:"name" validate:"alphaunicode"`
	}

	// Given
	testStruct := TestStructAlphaunicode{
		Name: "JohnDoé123",
	}

	// When
	errors := simba.ValidateStruct(testStruct)

	// Then
	assert.NotNil(t, errors)
	assert.Equal(t, 1, len(errors))
	assert.Equal(t, "name", errors[0].Field)
	assert.Equal(t, "name must contain only letters (no numbers allowed) that are part of unicode", errors[0].Err)

	// Given
	testStruct = TestStructAlphaunicode{
		Name: "JaneDoé",
	}

	// When
	errors = simba.ValidateStruct(testStruct)

	// Then
	assert.Nil(t, errors)
}

func TestValidateStruct_Numeric(t *testing.T) {
	t.Parallel()

	type TestStructNumeric struct {
		Age string `json:"age" validate:"numeric"`
	}

	// Given
	testStruct := TestStructNumeric{
		Age: "John",
	}

	// When
	errors := simba.ValidateStruct(testStruct)

	// Then
	assert.NotNil(t, errors)
	assert.Equal(t, 1, len(errors))
	assert.Equal(t, "age", errors[0].Field)
	assert.Equal(t, "age must be a numeric value", errors[0].Err)

	// Given
	testStruct = TestStructNumeric{
		Age: "1",
	}

	// When
	errors = simba.ValidateStruct(testStruct)

	// Then
	assert.Nil(t, errors)
}

// Format

func TestValidateStruct_Base64(t *testing.T) {
	t.Parallel()

	// Given
	type TestStructBase64 struct {
		Data string `json:"data" validate:"base64"`
	}
	testStruct := TestStructBase64{
		Data: "invalid-base64",
	}

	// When
	errors := simba.ValidateStruct(testStruct)

	// Then
	assert.NotNil(t, errors)
	assert.Equal(t, 1, len(errors))
	assert.Equal(t, "data", errors[0].Field)
	assert.Equal(t, "data must be a valid base64 encoded string", errors[0].Err)

	// Given
	testStruct = TestStructBase64{
		Data: "c29tZSB2YWxpZCBkYXRh",
	}

	// When
	errors = simba.ValidateStruct(testStruct)

	// Then
	assert.Nil(t, errors)
}

func TestValidateStruct_E164(t *testing.T) {
	t.Parallel()

	// Given
	type TestStructE164 struct {
		PhoneNumber string `json:"phoneNumber" validate:"e164"`
	}
	testStruct := TestStructE164{
		PhoneNumber: "invalid-phone",
	}

	// When
	errors := simba.ValidateStruct(testStruct)

	// Then
	assert.NotNil(t, errors)
	assert.Equal(t, 1, len(errors))
	assert.Equal(t, "phoneNumber", errors[0].Field)
	assert.Equal(t, "'invalid-phone' must be a valid E.164 formatted phone number", errors[0].Err)

	// Given
	testStruct = TestStructE164{
		PhoneNumber: "+14155552671",
	}

	// When
	errors = simba.ValidateStruct(testStruct)

	// Then
	assert.Nil(t, errors)
}

func TestValidateStruct_Email(t *testing.T) {
	t.Parallel()

	type TestStructEmail struct {
		Email string `json:"email" validate:"email"`
	}

	// Given
	testStruct := TestStructEmail{
		Email: "invalid-email",
	}

	// When
	errors := simba.ValidateStruct(testStruct)

	// Then
	assert.NotNil(t, errors)
	assert.Equal(t, 1, len(errors))
	assert.Equal(t, "email", errors[0].Field)
	assert.Equal(t, "'invalid-email' is not a valid email address", errors[0].Err)

	// Given
	testStruct = TestStructEmail{
		Email: "john.doe@example.com",
	}

	// When
	errors = simba.ValidateStruct(testStruct)

	// Then
	assert.Nil(t, errors)
}

func TestValidateStruct_Jwt(t *testing.T) {
	t.Parallel()

	// Given
	type TestStructJwt struct {
		Token string `json:"token" validate:"jwt"`
	}
	testStruct := TestStructJwt{
		Token: "invalid-jwt",
	}

	// When
	errors := simba.ValidateStruct(testStruct)

	// Then
	assert.NotNil(t, errors)
	assert.Equal(t, 1, len(errors))
	assert.Equal(t, "token", errors[0].Field)
	assert.Equal(t, "token must be a valid JWT token", errors[0].Err)

	// Given
	testStruct = TestStructJwt{
		Token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
	}

	// When
	errors = simba.ValidateStruct(testStruct)

	// Then
	assert.Nil(t, errors)
}

func TestValidateStruct_Uuid(t *testing.T) {
	t.Parallel()

	// Given
	type TestStructUuid struct {
		Id string `json:"id" validate:"uuid"`
	}
	testStruct := TestStructUuid{
		Id: "invalid-uuid",
	}

	// When
	errors := simba.ValidateStruct(testStruct)

	// Then
	assert.NotNil(t, errors)
	assert.Equal(t, 1, len(errors))
	assert.Equal(t, "id", errors[0].Field)
	assert.Equal(t, "id must be a valid UUID", errors[0].Err)

	// Given
	testStruct = TestStructUuid{
		Id: "550e8400-e29b-41d4-a716-446655440000",
	}

	// When
	errors = simba.ValidateStruct(testStruct)

	// Then
	assert.Nil(t, errors)
}

// Other

func TestValidateStruct_Len_String(t *testing.T) {
	t.Parallel()

	type TestStructLen struct {
		Name string `json:"name" validate:"len=3"`
	}

	// Given
	testStruct := TestStructLen{
		Name: "Jo",
	}

	// When
	errors := simba.ValidateStruct(testStruct)

	// Then
	assert.NotNil(t, errors)
	assert.Equal(t, 1, len(errors))
	assert.Equal(t, "name", errors[0].Field)
	assert.Equal(t, "name must be exactly 3 characters long", errors[0].Err)

	// Given
	testStruct = TestStructLen{
		Name: "Jan",
	}

	// When
	errors = simba.ValidateStruct(testStruct)

	// Then
	assert.Nil(t, errors)
}

func TestValidateStruct_Max_Int(t *testing.T) {
	t.Parallel()

	type TestStructMax struct {
		Age int `json:"age" validate:"max=130"`
	}

	// Given
	testStruct := TestStructMax{
		Age: 131,
	}

	// When
	errors := simba.ValidateStruct(testStruct)

	// Then
	assert.NotNil(t, errors)
	assert.Equal(t, 1, len(errors))
	assert.Equal(t, "age", errors[0].Field)
	assert.Equal(t, "age must not exceed 130", errors[0].Err)

	// Given
	testStruct = TestStructMax{
		Age: 130,
	}

	// When
	errors = simba.ValidateStruct(testStruct)

	// Then
	assert.Nil(t, errors)
}

func TestValidateStruct_Max_Float(t *testing.T) {
	t.Parallel()

	type TestStructMax struct {
		Price float64 `json:"price" validate:"max=100.34"`
	}

	// Given
	testStruct := TestStructMax{
		Price: 101,
	}

	// When
	errors := simba.ValidateStruct(testStruct)

	// Then
	assert.NotNil(t, errors)
	assert.Equal(t, 1, len(errors))
	assert.Equal(t, "price", errors[0].Field)
	assert.Equal(t, "price must not exceed 100.34", errors[0].Err)

	// Given
	testStruct = TestStructMax{
		Price: 100,
	}

	// When
	errors = simba.ValidateStruct(testStruct)

	// Then
	assert.Nil(t, errors)
}

func TestValidateStruct_Max_Slice(t *testing.T) {
	t.Parallel()

	type TestStructMax struct {
		Tags []string `json:"tags" validate:"max=3"`
	}

	// Given
	testStruct := TestStructMax{
		Tags: []string{"tag1", "tag2", "tag3", "tag4"},
	}

	// When
	errors := simba.ValidateStruct(testStruct)

	// Then
	assert.NotNil(t, errors)
	assert.Equal(t, 1, len(errors))
	assert.Equal(t, "tags", errors[0].Field)
	assert.Equal(t, "tags must not contain more than 3 items", errors[0].Err)

	// Given
	testStruct = TestStructMax{
		Tags: []string{"tag1", "tag2", "tag3"},
	}

	// When
	errors = simba.ValidateStruct(testStruct)

	// Then
	assert.Nil(t, errors)
}

func TestValidateStruct_Max_Map(t *testing.T) {
	t.Parallel()

	type TestStructMax struct {
		Meta map[string]string `json:"meta" validate:"max=1"`
	}

	// Given
	testStruct := TestStructMax{
		Meta: map[string]string{"key1": "value1", "key2": "value2"},
	}

	// When
	errors := simba.ValidateStruct(testStruct)

	// Then
	assert.NotNil(t, errors)
	assert.Equal(t, 1, len(errors))
	assert.Equal(t, "meta", errors[0].Field)
	assert.Equal(t, "meta must not contain more than 1 items", errors[0].Err)

	// Given
	testStruct = TestStructMax{
		Meta: map[string]string{"key1": "value1"},
	}

	// When
	errors = simba.ValidateStruct(testStruct)

	// Then
	assert.Nil(t, errors)
}

func TestValidateStruct_Max_String(t *testing.T) {
	t.Parallel()

	type TestStructMax struct {
		Name string `json:"name" validate:"max=2"`
	}

	// Given
	testStruct := TestStructMax{
		Name: "ABC",
	}

	// When
	errors := simba.ValidateStruct(testStruct)

	// Then
	assert.NotNil(t, errors)
	assert.Equal(t, 1, len(errors))
	assert.Equal(t, "name", errors[0].Field)
	assert.Equal(t, "name must not exceed 2 characters", errors[0].Err)

	// Given
	testStruct = TestStructMax{
		Name: "AB",
	}

	// When
	errors = simba.ValidateStruct(testStruct)

	// Then
	assert.Nil(t, errors)
}

func TestValidateStruct_Min_Int(t *testing.T) {
	t.Parallel()

	type TestStructMin struct {
		Age int `json:"age" validate:"min=0"`
	}

	// Given
	testStruct := TestStructMin{
		Age: -1,
	}

	// When
	errors := simba.ValidateStruct(testStruct)

	// Then
	assert.NotNil(t, errors)
	assert.Equal(t, 1, len(errors))
	assert.Equal(t, "age", errors[0].Field)
	assert.Equal(t, "age must be at least 0", errors[0].Err)

	// Given
	testStruct = TestStructMin{
		Age: 0,
	}

	// When
	errors = simba.ValidateStruct(testStruct)

	// Then
	assert.Nil(t, errors)
}

func TestValidateStruct_Min_Float(t *testing.T) {
	t.Parallel()

	type TestStructMin struct {
		Price float64 `json:"price" validate:"min=0.0"`
	}

	// Given
	testStruct := TestStructMin{
		Price: -1,
	}

	// When
	errors := simba.ValidateStruct(testStruct)

	// Then
	assert.NotNil(t, errors)
	assert.Equal(t, 1, len(errors))
	assert.Equal(t, "price", errors[0].Field)
	assert.Equal(t, "price must be at least 0.0", errors[0].Err)

	// Given
	testStruct = TestStructMin{
		Price: 0,
	}

	// When
	errors = simba.ValidateStruct(testStruct)

	// Then
	assert.Nil(t, errors)
}

func TestValidateStruct_Min_Slice(t *testing.T) {
	t.Parallel()

	type TestStructMin struct {
		Tags []string `json:"tags" validate:"min=2"`
	}

	// Given
	testStruct := TestStructMin{
		Tags: []string{"tag1"},
	}

	// When
	errors := simba.ValidateStruct(testStruct)

	// Then
	assert.NotNil(t, errors)
	assert.Equal(t, 1, len(errors))
	assert.Equal(t, "tags", errors[0].Field)
	assert.Equal(t, "tags must contain at least 2 items", errors[0].Err)

	// Given
	testStruct = TestStructMin{
		Tags: []string{"tag1", "tag2"},
	}

	// When
	errors = simba.ValidateStruct(testStruct)

	// Then
	assert.Nil(t, errors)
}

func TestValidateStruct_Min_Map(t *testing.T) {
	t.Parallel()

	type TestStructMin struct {
		Meta map[string]string `json:"meta" validate:"min=1"`
	}

	// Given
	testStruct := TestStructMin{
		Meta: map[string]string{},
	}

	// When
	errors := simba.ValidateStruct(testStruct)

	// Then
	assert.NotNil(t, errors)
	assert.Equal(t, 1, len(errors))
	assert.Equal(t, "meta", errors[0].Field)
	assert.Equal(t, "meta must contain at least 1 items", errors[0].Err)

	// Given
	testStruct = TestStructMin{
		Meta: map[string]string{"key": "value"},
	}

	// When
	errors = simba.ValidateStruct(testStruct)

	// Then
	assert.Nil(t, errors)
}

func TestValidateStruct_Min_String(t *testing.T) {
	t.Parallel()

	type TestStructMin struct {
		Name string `json:"name" validate:"min=2"`
	}

	// Given
	testStruct := TestStructMin{
		Name: "A",
	}

	// When
	errors := simba.ValidateStruct(testStruct)

	// Then
	assert.NotNil(t, errors)
	assert.Equal(t, 1, len(errors))
	assert.Equal(t, "name", errors[0].Field)
	assert.Equal(t, "name must be at least 2 characters long", errors[0].Err)

	// Given
	testStruct = TestStructMin{
		Name: "AB",
	}

	// When
	errors = simba.ValidateStruct(testStruct)

	// Then
	assert.Nil(t, errors)
}

func TestValidateStruct_Required(t *testing.T) {
	t.Parallel()

	type TestStructRequired struct {
		Name  string `json:"name" validate:"required"`
		Email string `json:"email" validate:"required"`
	}

	// Given
	testStruct := TestStructRequired{}

	// When
	errors := simba.ValidateStruct(testStruct)

	// Then
	assert.NotNil(t, errors)
	assert.Equal(t, 2, len(errors))
	assert.Equal(t, "name", errors[0].Field)
	assert.Equal(t, "name is required", errors[0].Err)
	assert.Equal(t, "email", errors[1].Field)
	assert.Equal(t, "email is required", errors[1].Err)

	// Given
	testStruct = TestStructRequired{
		Name:  "John Doe",
		Email: "john.doe@example.com",
	}

	// When
	errors = simba.ValidateStruct(testStruct)

	// Then
	assert.Nil(t, errors)
}

func TestValidateStruct_NoJsonTags(t *testing.T) {
	t.Parallel()

	type TestStructNoJsonTags struct {
		Name  string `validate:"required"`
		Email string `validate:"required"`
	}

	// Given
	testStruct := TestStructNoJsonTags{}

	// When
	errors := simba.ValidateStruct(testStruct)

	// Then
	assert.NotNil(t, errors)
	assert.Equal(t, 2, len(errors))
	assert.Equal(t, "Name", errors[0].Field)
	assert.Equal(t, "name is required", errors[0].Err)
	assert.Equal(t, "Email", errors[1].Field)
	assert.Equal(t, "email is required", errors[1].Err)

	// Given
	testStruct = TestStructNoJsonTags{
		Name:  "John Doe",
		Email: "john.doe@example.com",
	}

	// When
	errors = simba.ValidateStruct(testStruct)

	// Then
	assert.Nil(t, errors)
}
