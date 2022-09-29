package reflection_test

import (
	"fmt"
	"reflect"
	"testing"

	reflection "github.com/r-usenko/proto3-reflection"
	api "github.com/r-usenko/proto3-reflection/fixtures/gen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestParseServices(t *testing.T) {
	methods := reflection.ParseServices("api", []protoreflect.ExtensionType{
		api.E_Subject,
		api.E_Consumer,
		api.E_Stream,
		api.E_IsStreamTransport,
	}, []protoreflect.ExtensionType{
		api.E_Reply,
		api.E_Subscribe,
		api.E_SubscribeQueue,
	})

	tests := map[string]map[string]interface{}{
		"Service1.Method11.Reply": {
			"Subject":           "players.player.create",
			"IsStreamTransport": true,
		},
		"Service1.Method11.Subscribe": {
			"Subject":           "players.player.delete",
			"IsStreamTransport": false,
		},
		"Service2.Method22.SubscribeQueue": {
			"Stream":   "ANTI_FRAUD",
			"Subject":  "players.player.update",
			"Consumer": "CABINET_REVIEW_STATUS_UPDATE",
		},
		"Service2.Method23.SubscribeQueue": {
			"Stream":   "ANTI_FRAUD",
			"Subject":  "players.player.update",
			"Consumer": "CABINET_REVIEW_STATUS_UPDATE",
		},
	}

	for methodFullName, info := range methods {
		t.Logf("Method: %q\n", methodFullName)

		for scenario, options := range info.Scenarios() {
			key := fmt.Sprintf("%s.%s.%s",
				info.ServiceDescriptor().Name(),
				info.MethodDescriptor().Name(),
				scenario.TypeDescriptor().Name(),
			)

			var optionsMap = make(map[string]interface{})
			for k, v := range options {
				optionsMap[string(k.TypeDescriptor().Name())] = v
			}

			if !reflect.DeepEqual(tests[key], optionsMap) {
				t.Errorf("FAILED : %q (%#v != %#v)\n", key, optionsMap, tests[key])
			}
		}
	}
}
