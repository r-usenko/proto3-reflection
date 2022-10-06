package reflection

import (
	"context"
	"fmt"
	"reflect"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

var ctxType = reflect.TypeOf((*context.Context)(nil)).Elem()
var protoType = reflect.TypeOf((*proto.Message)(nil)).Elem()
var errorType = reflect.TypeOf((*error)(nil)).Elem()

type MethodInfo struct {
	Service protoreflect.ServiceDescriptor
	Method  protoreflect.MethodDescriptor

	//map[scenario]map[options]extensionValue
	Scenarios map[protoreflect.ExtensionType]map[protoreflect.ExtensionType]interface{}
}

type ImplementationInfo struct {
	MethodValue        reflect.Value
	RequestMessageType reflect.Type
}

// ParseImplementation
//
//	{package}.{serviceName}:*implementationStruct
func ParseImplementation(services map[string]interface{}) map[string]ImplementationInfo {
	var res = make(map[string]ImplementationInfo)

	for serviceName, implementation := range services {
		serviceValue := reflect.ValueOf(implementation)

		if !serviceValue.IsValid() || serviceValue.IsZero() {
			//skip invalid implementation
			continue
		}

		serviceType := serviceValue.Type()
		if serviceType.Kind() != reflect.Struct && serviceType.Kind() != reflect.Pointer {
			//skip invalid implementation
			continue
		}

		for i := 0; i < serviceValue.NumMethod(); i++ {
			methodValue := serviceValue.Method(i)
			methodType := methodValue.Type()

			//func(context.Context, proto.Message) (proto.Message, error)
			if methodType.NumIn() != 2 ||
				methodType.NumOut() != 2 ||
				!methodType.In(0).ConvertibleTo(ctxType) ||
				!methodType.In(1).ConvertibleTo(protoType) ||
				!methodType.Out(0).ConvertibleTo(protoType) ||
				!methodType.Out(1).ConvertibleTo(errorType) {

				//skip not proto method
				continue
			}

			res[fmt.Sprintf("%s.%s", serviceName, serviceType.Method(i).Name)] = ImplementationInfo{
				MethodValue:        methodValue,
				RequestMessageType: methodType.In(1),
			}
		}
	}

	return res
}

// ParseProtoServices
//
//	 package {$packageName}
//		extend google.protobuf.EnumValueOptions []{enumOptions}
//		extend google.protobuf.MethodOptions []{methodOptions}
func ParseProtoServices(packageName protoreflect.FullName, enumOptions []protoreflect.ExtensionType, methodOptions []protoreflect.ExtensionType) map[protoreflect.FullName]MethodInfo {
	result := make(map[protoreflect.FullName]MethodInfo)

	protoregistry.GlobalFiles.RangeFilesByPackage(packageName, func(descriptor protoreflect.FileDescriptor) bool {
		for sI := 0; sI < descriptor.Services().Len(); sI++ {
			service := descriptor.Services().Get(sI)

			for mI := 0; mI < service.Methods().Len(); mI++ {
				method := service.Methods().Get(mI)
				mi := MethodInfo{
					Method:    method,
					Service:   service,
					Scenarios: make(map[protoreflect.ExtensionType]map[protoreflect.ExtensionType]interface{}),
				}

				for _, enum := range methodOptions {
					if !proto.HasExtension(method.Options(), enum) {
						continue
					}

					enumVal, ok := proto.GetExtension(method.Options(), enum).(protoreflect.Enum)
					if !ok {
						continue
					}

					mi.Scenarios[enum] = make(map[protoreflect.ExtensionType]interface{})
					enumValOptions := enumVal.Descriptor().Values().ByNumber(enumVal.Number()).Options()

					for _, opt := range enumOptions {
						if !proto.HasExtension(enumValOptions, opt) {
							continue
						}

						mi.Scenarios[enum][opt] = proto.GetExtension(enumValOptions, opt)
					}
				}

				result[method.FullName()] = mi
			}
		}

		return true
	})

	return result
}
