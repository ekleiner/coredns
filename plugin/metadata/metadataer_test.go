package metadata

import (
	"context"
	"reflect"
	"testing"
)

func TestMD(t *testing.T) {
	tests := []struct {
		addValues      map[string]interface{}
		expectedValues map[string]interface{}
	}{
		{
			// Add initial metadata key/vals
			map[string]interface{}{"key1": "val1", "key2": 2, "key3": 3},
			map[string]interface{}{"key1": "val1", "key2": 2, "key3": 3},
		},
		{
			// Add additional key/vals. Duplicate keys are removed.
			map[string]interface{}{"key2": 2, "key3": 3, "key4": 4},
			map[string]interface{}{"key1": "val1", "key4": 4},
		},
	}

	// Using one same md and ctx for all test cases
	ctx := context.TODO()
	md, ctx := newMD(ctx)

	for i, tc := range tests {
		//
		md.addValues(tc.addValues)

		if !reflect.DeepEqual(tc.expectedValues, map[string]interface{}(md)) {
			t.Errorf("Test %d: Expected %v but got %v", i, tc.expectedValues, md)
		}

		// Make sure that MD is recieved from context successfullly
		mdFromContext, ok := FromContext(ctx)
		if !ok {
			t.Errorf("Test %d: MD is not recieved from the context", i)
		}
		if !reflect.DeepEqual(md, mdFromContext) {
			t.Errorf("Test %d: MD recieved from context differs from initial. Initial: %v, from context: %v", i, md, mdFromContext)
		}
	}
}
