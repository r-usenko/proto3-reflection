package reflection_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	reflection "github.com/r-usenko/proto3-reflection"
	api "github.com/r-usenko/proto3-reflection/fixtures/gen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestParseServices(t *testing.T) {
	methods := reflection.ParseProtoServices("api", []protoreflect.ExtensionType{
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

		for scenario, options := range info.Scenarios {
			key := fmt.Sprintf("%s.%s.%s",
				info.Service.Name(),
				info.Method.Name(),
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

func TestParseImplementation(t *testing.T) {
	s := new(implementation)

	m := reflection.ParseImplementation(map[string]interface{}{
		"api.Service1": s.UnimplementedService1Server,
		"api.Service2": s.UnimplementedService2Server,
		"api":          s,
		"api2":         implementation{},
		"nil":          nil,
		"string":       "text",
		"interface":    new(api.Service1Server),
	})

	methods := make(map[string]struct{})
	for k := range m {
		methods[k] = struct{}{}
	}

	if !reflect.DeepEqual(methods, map[string]struct{}{
		"api.Method11":             {},
		"api.Method21":             {},
		"api.Method22":             {},
		"api.Method23":             {},
		"api.UndefinedProtoMethod": {},
	}) {
		t.Errorf("FAILED: Result are different %v\n", methods)
	}
}

type implementation struct {
	api.UnimplementedService1Server
	api.UnimplementedService2Server
}

func (m *implementation) Method21(ctx context.Context, request1 *api.Request1) (*api.Response1, error) {
	//TODO implement me
	panic("implement me")
}

func (m *implementation) Method22(ctx context.Context, request1 *api.Request1) (*api.Response1, error) {
	//TODO implement me
	panic("implement me")
}

func (m *implementation) Method23(ctx context.Context, request2 *api.Request2) (*api.Response2, error) {
	//TODO implement me
	panic("implement me")
}

func (m *implementation) Method11(ctx context.Context, request1 *api.Request1) (*api.Response1, error) {
	//TODO implement me
	panic("implement me")
}

func (m *implementation) UndefinedProtoMethod(ctx context.Context, request1 *api.Request1) (*api.Response1, error) {
	//TODO implement me
	panic("implement me")
}

func (m *implementation) hiddenMethod(ctx context.Context, request1 *api.Request1) (*api.Response1, error) {
	//TODO implement me
	panic("implement me")
}

func (m *implementation) AnotherMethod() {
	//TODO implement me
	panic("implement me")
}

var _ api.Service1Server = (*implementation)(nil)
var _ api.Service2Server = (*implementation)(nil)
