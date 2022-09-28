package reflection

import (
	"context"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

type Callback func(context.Context, proto.Message) (proto.Message, error)

type methodInfo struct {
	service   protoreflect.ServiceDescriptor
	method    protoreflect.MethodDescriptor
	scenarios map[protoreflect.ExtensionType]map[protoreflect.ExtensionType]string
}

func (m *methodInfo) MethodDescriptor() protoreflect.MethodDescriptor {
	return m.method
}

func (m *methodInfo) ServiceDescriptor() protoreflect.ServiceDescriptor {
	return m.service
}

// Scenarios
//
//	map[scenario]map[options]
func (m *methodInfo) Scenarios() map[protoreflect.ExtensionType]map[protoreflect.ExtensionType]string {
	return m.scenarios
}

// GenerateIOMessage Request/Response from method
func GenerateIOMessage(method protoreflect.MethodDescriptor) (proto.Message, proto.Message, error) {
	var msgPI, msgPO protoreflect.MessageType
	var err error

	msgPI, err = protoregistry.GlobalTypes.FindMessageByName(method.Input().FullName())
	if err != nil {
		return nil, nil, err
	}
	msgPO, err = protoregistry.GlobalTypes.FindMessageByName(method.Output().FullName())
	if err != nil {
		return nil, nil, err
	}

	return msgPI.New().Interface(), msgPO.New().Interface(), nil
}

// ParseServices
//
//	 package {$packageName}
//		extend google.protobuf.EnumValueOptions []{enumOptions}
//		extend google.protobuf.MethodOptions []{methodOptions}
func ParseServices(packageName protoreflect.FullName, enumOptions []protoreflect.ExtensionType, methodOptions []protoreflect.ExtensionType) map[protoreflect.FullName]methodInfo {
	result := make(map[protoreflect.FullName]methodInfo)

	protoregistry.GlobalFiles.RangeFilesByPackage(packageName, func(descriptor protoreflect.FileDescriptor) bool {
		for sI := 0; sI < descriptor.Services().Len(); sI++ {
			service := descriptor.Services().Get(sI)

			for mI := 0; mI < service.Methods().Len(); mI++ {
				method := service.Methods().Get(mI)
				mi := methodInfo{
					method:    method,
					service:   service,
					scenarios: make(map[protoreflect.ExtensionType]map[protoreflect.ExtensionType]string),
				}

				for _, enum := range methodOptions {
					if !proto.HasExtension(method.Options(), enum) {
						continue
					}

					enumVal, ok := proto.GetExtension(method.Options(), enum).(protoreflect.Enum)
					if !ok {
						continue
					}

					mi.scenarios[enum] = make(map[protoreflect.ExtensionType]string)
					enumValOptions := enumVal.Descriptor().Values().ByNumber(enumVal.Number()).Options()

					for _, opt := range enumOptions {
						if !proto.HasExtension(enumValOptions, opt) {
							continue
						}

						mi.scenarios[enum][opt] = proto.GetExtension(enumValOptions, opt).(string)
					}
				}

				result[method.FullName()] = mi
			}
		}

		return true
	})

	return result
}
