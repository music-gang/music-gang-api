package common_test

import (
	"testing"

	"github.com/music-gang/music-gang-api/common"
)

func TestBytes(t *testing.T) {
	t.Run("OK - map", func(t *testing.T) {

		m := map[string]any{
			"a": true,
			"b": 34,
			"c": "hello",
		}

		b, err := common.ToBytes(m)
		if err != nil {
			t.Errorf("Error while converting map to bytes: %s", err.Error())
		}

		m2 := make(map[string]any)

		if err := common.FromBytes(b, &m2); err != nil {
			t.Errorf("Error while converting bytes to map: %s", err.Error())
		}

		if _, ok := m2["a"]; !ok || m2["a"] != true {
			t.Errorf("Map not converted correctly")
		}

		if _, ok := m2["b"]; !ok || m2["b"] != 34 {
			t.Errorf("Map not converted correctly")
		}

		if _, ok := m2["c"]; !ok || m2["c"] != "hello" {
			t.Errorf("Map not converted correctly")
		}
	})
}
