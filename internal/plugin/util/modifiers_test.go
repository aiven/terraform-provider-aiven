package util

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFlattenObjectsToArray(t *testing.T) {
	t.Run("basic flatten", func(t *testing.T) {
		r := NewRawMap([]byte(`{"tech_emails": [{"email": "a@aiven.io"}, {"email": "b@aiven.io"}]}`))
		err := FlattenObjectsToArray[any]("email", "tech_emails")(r, nil)
		require.NoError(t, err)

		expected := `{"tech_emails":["a@aiven.io","b@aiven.io"]}`
		require.JSONEq(t, expected, string(r.GetData()))
	})

	t.Run("nested path", func(t *testing.T) {
		r := NewRawMap([]byte(`{"data": {"tech_emails": [{"email": "a@aiven.io"}, {"email": "b@aiven.io"}]}}`))
		err := FlattenObjectsToArray[any]("email", "data", "tech_emails")(r, nil)
		require.NoError(t, err)

		expected := `{"data":{"tech_emails":["a@aiven.io","b@aiven.io"]}}`
		require.JSONEq(t, expected, string(r.GetData()))
	})

	t.Run("non-existent path", func(t *testing.T) {
		r := NewRawMap([]byte(`{}`))
		err := FlattenObjectsToArray[any]("email", "tech_emails")(r, nil)
		require.NoError(t, err)

		expected := `{}`
		require.JSONEq(t, expected, string(r.GetData()))
	})

	t.Run("empty array", func(t *testing.T) {
		r := NewRawMap([]byte(`{"tech_emails": []}`))
		err := FlattenObjectsToArray[any]("email", "tech_emails")(r, nil)
		require.NoError(t, err)

		expected := `{"tech_emails":[]}`
		require.JSONEq(t, expected, string(r.GetData()))
	})

	t.Run("complex objects", func(t *testing.T) {
		r := NewRawMap([]byte(`{"users": [{"id": 1, "name": "John"}, {"id": 2, "name": "Jane"}]}`))
		err := FlattenObjectsToArray[any]("name", "users")(r, nil)
		require.NoError(t, err)

		expected := `{"users":["John","Jane"]}`
		require.JSONEq(t, expected, string(r.GetData()))
	})
}

func TestExpandArrayToObjects(t *testing.T) {
	t.Run("basic expand", func(t *testing.T) {
		r := NewRawMap([]byte(`{"tech_emails": ["a@aiven.io", "b@aiven.io"]}`))
		err := ExpandArrayToObjects[any](false, "email", "tech_emails")(r, nil)
		require.NoError(t, err)

		expected := `{"tech_emails":[{"email":"a@aiven.io"},{"email":"b@aiven.io"}]}`
		require.JSONEq(t, expected, string(r.GetData()))
	})

	t.Run("nested path", func(t *testing.T) {
		r := NewRawMap([]byte(`{"data": {"tech_emails": ["a@aiven.io", "b@aiven.io"]}}`))
		err := ExpandArrayToObjects[any](false, "email", "data", "tech_emails")(r, nil)
		require.NoError(t, err)

		expected := `{"data":{"tech_emails":[{"email":"a@aiven.io"},{"email":"b@aiven.io"}]}}`
		require.JSONEq(t, expected, string(r.GetData()))
	})

	t.Run("non-existent path not required", func(t *testing.T) {
		r := NewRawMap([]byte(`{}`))
		err := ExpandArrayToObjects[any](false, "email", "tech_emails")(r, nil)
		require.NoError(t, err)

		expected := `{}`
		require.JSONEq(t, expected, string(r.GetData()))
	})

	t.Run("non-existent path required", func(t *testing.T) {
		r := NewRawMap([]byte(`{}`))
		err := ExpandArrayToObjects[any](true, "email", "tech_emails")(r, nil)
		require.NoError(t, err)

		expected := `{"tech_emails":[]}`
		require.JSONEq(t, expected, string(r.GetData()))
	})

	t.Run("empty array", func(t *testing.T) {
		r := NewRawMap([]byte(`{"tech_emails": []}`))
		err := ExpandArrayToObjects[any](false, "email", "tech_emails")(r, nil)
		require.NoError(t, err)

		expected := `{"tech_emails":[]}`
		require.JSONEq(t, expected, string(r.GetData()))
	})

	t.Run("mixed types", func(t *testing.T) {
		r := NewRawMap([]byte(`{"values": ["string", 123, true]}`))
		err := ExpandArrayToObjects[any](false, "value", "values")(r, nil)
		require.NoError(t, err)

		expected := `{"values":[{"value":"string"},{"value":123},{"value":true}]}`
		require.JSONEq(t, expected, string(r.GetData()))
	})

	t.Run("numeric values", func(t *testing.T) {
		r := NewRawMap([]byte(`{"ids": [1, 2, 3]}`))
		err := ExpandArrayToObjects[any](false, "id", "ids")(r, nil)
		require.NoError(t, err)

		expected := `{"ids":[{"id":1},{"id":2},{"id":3}]}`
		require.JSONEq(t, expected, string(r.GetData()))
	})

	t.Run("null value not required", func(t *testing.T) {
		r := NewRawMap([]byte(`{"tech_emails": null}`))
		err := ExpandArrayToObjects[any](false, "email", "tech_emails")(r, nil)
		require.NoError(t, err)

		expected := `{"tech_emails":[]}`
		require.JSONEq(t, expected, string(r.GetData()))
	})

	t.Run("null value required", func(t *testing.T) {
		r := NewRawMap([]byte(`{"tech_emails": null}`))
		err := ExpandArrayToObjects[any](true, "email", "tech_emails")(r, nil)
		require.NoError(t, err)

		expected := `{"tech_emails":[]}`
		require.JSONEq(t, expected, string(r.GetData()))
	})

	t.Run("invalid type", func(t *testing.T) {
		r := NewRawMap([]byte(`{"tech_emails": "not an array"}`))
		err := ExpandArrayToObjects[any](false, "email", "tech_emails")(r, nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "expected array")
	})
}

func TestFlattenMapToKeyValue(t *testing.T) {
	t.Run("basic flatten", func(t *testing.T) {
		r := NewRawMap([]byte(`{"tags": {"environment": "prod", "role": "admin"}}`))
		err := FlattenMapToKeyValue[any]("key", "value", "tags")(r, nil)
		require.NoError(t, err)

		// Note: Order of elements in the array depends on map iteration order
		expected := `{"tags":[{"key":"environment","value":"prod"},{"key":"role","value":"admin"}]}`
		require.JSONEq(t, expected, string(r.GetData()))
	})

	t.Run("nested path", func(t *testing.T) {
		r := NewRawMap([]byte(`{"data": {"tags": {"environment": "prod"}}}`))
		err := FlattenMapToKeyValue[any]("key", "value", "data", "tags")(r, nil)
		require.NoError(t, err)

		expected := `{"data":{"tags":[{"key":"environment","value":"prod"}]}}`
		require.JSONEq(t, expected, string(r.GetData()))
	})

	t.Run("non-existent path", func(t *testing.T) {
		r := NewRawMap([]byte(`{}`))
		err := FlattenMapToKeyValue[any]("key", "value", "tags")(r, nil)
		require.NoError(t, err)

		expected := `{}`
		require.JSONEq(t, expected, string(r.GetData()))
	})

	t.Run("empty map", func(t *testing.T) {
		r := NewRawMap([]byte(`{"tags": {}}`))
		err := FlattenMapToKeyValue[any]("key", "value", "tags")(r, nil)
		require.NoError(t, err)

		expected := `{"tags":[]}`
		require.JSONEq(t, expected, string(r.GetData()))
	})

	t.Run("single entry", func(t *testing.T) {
		r := NewRawMap([]byte(`{"tags": {"foo": "bar"}}`))
		err := FlattenMapToKeyValue[any]("key", "value", "tags")(r, nil)
		require.NoError(t, err)

		expected := `{"tags":[{"key":"foo","value":"bar"}]}`
		require.JSONEq(t, expected, string(r.GetData()))
	})
}

func TestExpandKeyValueToMap(t *testing.T) {
	t.Run("basic expand", func(t *testing.T) {
		r := NewRawMap([]byte(`{"tags": [{"key": "environment", "value": "prod"}, {"key": "role", "value": "admin"}]}`))
		err := ExpandKeyValueToMap[any](false, "key", "value", "tags")(r, nil)
		require.NoError(t, err)

		expected := `{"tags":{"environment":"prod","role":"admin"}}`
		require.JSONEq(t, expected, string(r.GetData()))
	})

	t.Run("nested path", func(t *testing.T) {
		r := NewRawMap([]byte(`{"data": {"tags": [{"key": "environment", "value": "prod"}]}}`))
		err := ExpandKeyValueToMap[any](false, "key", "value", "data", "tags")(r, nil)
		require.NoError(t, err)

		expected := `{"data":{"tags":{"environment":"prod"}}}`
		require.JSONEq(t, expected, string(r.GetData()))
	})

	t.Run("non-existent path not required", func(t *testing.T) {
		r := NewRawMap([]byte(`{}`))
		err := ExpandKeyValueToMap[any](false, "key", "value", "tags")(r, nil)
		require.NoError(t, err)

		expected := `{}`
		require.JSONEq(t, expected, string(r.GetData()))
	})

	t.Run("non-existent path required", func(t *testing.T) {
		r := NewRawMap([]byte(`{}`))
		err := ExpandKeyValueToMap[any](true, "key", "value", "tags")(r, nil)
		require.NoError(t, err)

		expected := `{"tags":{}}`
		require.JSONEq(t, expected, string(r.GetData()))
	})

	t.Run("empty array", func(t *testing.T) {
		r := NewRawMap([]byte(`{"tags": []}`))
		err := ExpandKeyValueToMap[any](false, "key", "value", "tags")(r, nil)
		require.NoError(t, err)

		expected := `{"tags":{}}`
		require.JSONEq(t, expected, string(r.GetData()))
	})

	t.Run("single entry", func(t *testing.T) {
		r := NewRawMap([]byte(`{"tags": [{"key": "foo", "value": "bar"}]}`))
		err := ExpandKeyValueToMap[any](false, "key", "value", "tags")(r, nil)
		require.NoError(t, err)

		expected := `{"tags":{"foo":"bar"}}`
		require.JSONEq(t, expected, string(r.GetData()))
	})

	t.Run("null value not required", func(t *testing.T) {
		r := NewRawMap([]byte(`{"tags": null}`))
		err := ExpandKeyValueToMap[any](false, "key", "value", "tags")(r, nil)
		require.NoError(t, err)

		expected := `{"tags":{}}`
		require.JSONEq(t, expected, string(r.GetData()))
	})

	t.Run("null value required", func(t *testing.T) {
		r := NewRawMap([]byte(`{"tags": null}`))
		err := ExpandKeyValueToMap[any](true, "key", "value", "tags")(r, nil)
		require.NoError(t, err)

		expected := `{"tags":{}}`
		require.JSONEq(t, expected, string(r.GetData()))
	})

	t.Run("invalid type", func(t *testing.T) {
		r := NewRawMap([]byte(`{"tags": "not an array"}`))
		err := ExpandKeyValueToMap[any](false, "key", "value", "tags")(r, nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "expected array")
	})
}

func TestComposeModifiers(t *testing.T) {
	t.Run("compose multiple modifiers", func(t *testing.T) {
		r := NewRawMap([]byte(`{
			"tech_emails": [{"email": "a@aiven.io"}, {"email": "b@aiven.io"}],
			"tags": {"environment": "prod", "role": "admin"}
		}`))

		modifier := ComposeModifiers(
			FlattenObjectsToArray[any]("email", "tech_emails"),
			FlattenMapToKeyValue[any]("key", "value", "tags"),
		)

		err := modifier(r, nil)
		require.NoError(t, err)

		// Note: Order of tags array elements depends on map iteration order
		expected := `{
			"tech_emails": ["a@aiven.io", "b@aiven.io"],
			"tags": [{"key":"environment","value":"prod"},{"key":"role","value":"admin"}]
		}`
		require.JSONEq(t, expected, string(r.GetData()))
	})

	t.Run("compose no modifiers", func(t *testing.T) {
		r := NewRawMap([]byte(`{"key": "value"}`))
		modifier := ComposeModifiers[any]()
		err := modifier(r, nil)
		require.NoError(t, err)

		expected := `{"key":"value"}`
		require.JSONEq(t, expected, string(r.GetData()))
	})

	t.Run("compose single modifier", func(t *testing.T) {
		r := NewRawMap([]byte(`{"tech_emails": [{"email": "a@aiven.io"}]}`))
		modifier := ComposeModifiers(FlattenObjectsToArray[any]("email", "tech_emails"))
		err := modifier(r, nil)
		require.NoError(t, err)

		expected := `{"tech_emails":["a@aiven.io"]}`
		require.JSONEq(t, expected, string(r.GetData()))
	})

	t.Run("compose with chained transformations", func(t *testing.T) {
		// Flatten -> Expand -> Flatten cycle
		r := NewRawMap([]byte(`{"data": [{"value": "a"}, {"value": "b"}]}`))

		modifier := ComposeModifiers(
			FlattenObjectsToArray[any]("value", "data"),
			ExpandArrayToObjects[any](false, "item", "data"),
		)

		err := modifier(r, nil)
		require.NoError(t, err)

		expected := `{"data":[{"item":"a"},{"item":"b"}]}`
		require.JSONEq(t, expected, string(r.GetData()))
	})
}
