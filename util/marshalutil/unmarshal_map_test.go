package marshalutil_test

import (
	"testing"

	"github.com/matthiasharzer/livebuffer/util/marshalutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnmarshalAny(t *testing.T) {
	t.Run("unmarshal map to struct", func(t *testing.T) {
		data := map[string]any{
			"name": "John",
			"age":  30,
		}

		type Person struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}

		var person Person
		err := marshalutil.UnmarshalAny(data, &person)
		require.NoError(t, err)

		assert.Equal(t, "John", person.Name)
		assert.Equal(t, 30, person.Age)
	})

	t.Run("unmarshal struct to map", func(t *testing.T) {
		type Person struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}

		person := Person{
			Name: "Alice",
			Age:  25,
		}

		var data map[string]any
		err := marshalutil.UnmarshalAny(person, &data)
		require.NoError(t, err)

		assert.Equal(t, "Alice", data["name"])
		assert.Equal(t, float64(25), data["age"]) // JSON numbers are unmarshaled as float64
	})

	t.Run("unmarshal map to map", func(t *testing.T) {
		data := map[string]any{
			"key1": "value1",
			"key2": 42,
		}

		var result map[string]any
		err := marshalutil.UnmarshalAny(data, &result)
		require.NoError(t, err)

		assert.Equal(t, "value1", result["key1"])
		assert.Equal(t, float64(42), result["key2"]) // JSON numbers are unmarshaled as float64
	})

	t.Run("unmarshal struct to struct", func(t *testing.T) {
		type Person struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}

		person := Person{
			Name: "Bob",
			Age:  40,
		}

		var result Person
		err := marshalutil.UnmarshalAny(person, &result)
		require.NoError(t, err)

		assert.Equal(t, "Bob", result.Name)
		assert.Equal(t, 40, result.Age)
	})

	t.Run("unmarshal map to struct with missing fields", func(t *testing.T) {
		data := map[string]any{
			"name": "Charlie",
		}

		type Person struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}

		var person Person
		err := marshalutil.UnmarshalAny(data, &person)
		require.NoError(t, err)

		assert.Equal(t, "Charlie", person.Name)
		assert.Equal(t, 0, person.Age) // Age should be zero value since it's missing in the map
	})
}
