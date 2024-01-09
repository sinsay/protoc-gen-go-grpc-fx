package main

import (
	"fmt"
	"os"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

type ServiceInfo struct {
	ServiceName     string
	ServiceFullName string
	ServiceNumber   int
}

// getRegistryName find service register name by location and option
func getRegistryName(gen *protogen.Plugin, service *protogen.Service) string {
	var pb *descriptorpb.FileDescriptorProto = nil
	for _, f := range gen.Request.ProtoFile {
		// api/ping-service/v1/services/ping.service.v1.proto
		if f.Name != nil && *f.Name == service.Location.SourceFile {
			pb = f
			break
		}
	}

	if pb == nil {
		return "CANT FIND SOURCE FILE FOR " + service.GoName
	}

	var srv *descriptorpb.ServiceDescriptorProto = nil
	for _, f := range pb.Service {
		if f.Name != nil && *f.Name == service.GoName {
			srv = f
			break
		}
	}

	if srv == nil {
		return "CANT FIND SERVICE FOR " + service.GoName
	}

	options := srv.GetOptions()
	if options == nil {
		return "HAVEN'T SET OPTION OF SERVICE NAME FOR " + service.GoName
	}

	serviceInfos := make([]ServiceInfo, 0)

	options.ProtoReflect().Range(func(fs protoreflect.FieldDescriptor, value protoreflect.Value) bool {
		serviceNumber := int(fs.Number())             // 99999
		serviceFiledName := fs.Name()                 // name some_name
		serviceFieldFullName := string(fs.FullName()) // api.some.service.some_service.some_name
		foundServiceName := value.String()

		// start with package ends with service name
		if strings.Index(string(serviceFieldFullName), strings.ToLower(*pb.Package)) == 0 &&
			strings.Index(string(serviceFiledName), strings.ToLower(srv.GetName())) == 0 {

			serviceInfos = append(serviceInfos, ServiceInfo{
				ServiceName:     foundServiceName,
				ServiceFullName: serviceFieldFullName,
				ServiceNumber:   serviceNumber,
			})
		}

		return true
	})

	if len(serviceInfos) > 0 {
		if len(serviceInfos) > 1 {
			_, _ = fmt.Fprintf(os.Stderr, "[Service] [WARNING] Got multiple service name!\n")
		}
		for _, s := range serviceInfos {
			_, _ = fmt.Fprintf(os.Stderr, "[Service] Got GRPC Service [%s] with name %s:%d\n", s.ServiceFullName, s.ServiceName, s.ServiceNumber)
		}

		return serviceInfos[0].ServiceName
	}

	return "DOESN'T MATCH OPTION STRING FOR " + service.GoName
}

func GenerateClientPureInjection(clientName string, g *protogen.GeneratedFile, service *protogen.Service, gen *protogen.Plugin) {
	regName := getRegistryName(gen, service)
	if len(regName) == 0 {
		return
	}

	if regName[0] != '/' {
		regName = "/" + regName
	}

	//	// Generate Service Name
	g.P("func (c *", unexport(service.GoName), "Client) RegisterNameForDiscover() string {")
	g.P("return \"", regName, "\"")
	g.P("}")
	g.P()

	g.P("func New", clientName, " (cc ", grpcPackage.Ident("ClientConnInterface"), ") ", clientName, " {")
	helper.generateNewClientDefinitions(g, service, clientName)
	g.P("}")
	g.P()

	// Name Provider
	g.P("func register", clientName, "GRPCNameProvider() []string {")
	g.P("    return []string {\"", regName, "\", \"grpc\"}")
	g.P("}")

	g.P("func Register", clientName, "GRPCProvider(creator interface{}) []interface{} {")
	g.P("    return []interface{} {")
	g.P("        fx.Annotate(")
	g.P("            New", clientName, ",")
	g.P("            fx.As(new(", clientName, ")),")
	g.P("            fx.ParamTags(`name:\"", regName, "/grpc/", unexport(service.GoName), "\"`),")
	g.P("         ),")
	g.P("        fx.Annotate(")
	g.P("            creator,")
	g.P("            fx.As(new(grpc.ClientConnInterface)),")
	g.P("            fx.ParamTags(`name:\"", regName, "/grpc/name/", unexport(service.GoName), "\"`),")
	g.P("            fx.ResultTags(`name:\"", regName, "/grpc/", unexport(service.GoName), "\"`),")
	g.P("         ),")
	g.P("         fx.Annotate(")
	g.P("            register", clientName, "GRPCNameProvider,")
	g.P("            fx.ResultTags(`name:\"", regName, "/grpc/name/", unexport(service.GoName), "\"`),")
	g.P("         ),")
	g.P("    }")
	g.P("}")
	g.P()

}
