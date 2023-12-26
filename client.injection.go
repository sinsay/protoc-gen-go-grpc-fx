package main

import (
	"fmt"
	"regexp"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/descriptorpb"
)

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
	// [api.ping.service.pingservicev1.name]:"permission-service"
	// check and extract name from string
	regText := fmt.Sprintf(`\[%s.name\]:\"(.+?)\"`, *pb.Package)
	reg := regexp.MustCompile(regText)
	extractFormula := reg.FindStringSubmatch(options.String())
	if len(extractFormula) <= 1 {
		return "DOESN'T MATCH OPTION STRING FOR " + service.GoName
	}

	return extractFormula[1]
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
	g.P("            fx.ParamTags(`name:\"", regName, "/grpc\"`),")
	g.P("         ),")
	g.P("        fx.Annotate(")
	g.P("            creator,")
	g.P("            fx.As(new(grpc.ClientConnInterface)),")
	g.P("            fx.ParamTags(`name:\"", regName, "/grpc/name\"`),")
	g.P("            fx.ResultTags(`name:\"", regName, "/grpc\"`),")
	g.P("         ),")
	g.P("         fx.Annotate(")
	g.P("            register", clientName, "GRPCNameProvider,")
	g.P("            fx.ResultTags(`name:\"", regName, "/grpc/name\"`),")
	g.P("         ),")
	g.P("    }")
	g.P("}")
	g.P()

}

//func GenerateClientInjection(clientName string, g *protogen.GeneratedFile, service *protogen.Service, gen *protogen.Plugin) {
//
//	g.P("func New", clientName, " (cc ", grpcPackage.Ident("ClientConnInterface"), ") ", clientName, " {")
//	helper.generateNewClientDefinitions(g, service, clientName)
//	g.P("}")
//	g.P()
//
//	// Generate Service Name
//	// try to get service name
//	g.P("func (c *", unexport(service.GoName), "Client) RegisterNameForDiscover() string {")
//	regName := getRegistryName(gen, service)
//	g.P("return \"", regName, "\"")
//	g.P("}")
//	g.P()
//
//	// generate init function
//	g.P("func init() {")
//	g.P("injection.InjectMany(Register", clientName, "Provider()...)")
//	g.P("}")
//
//	// generate method for create instance
//	g.P("func Register", clientName, "Provider() []interface{} {")
//	g.P("return []interface{}{")
//	g.P("	fx.Annotate(")
//	g.P("		New", clientName, ",")
//	g.P("		fx.As(new(", clientName, ")),")
//	g.P("		fx.ParamTags(`name:\"", regName, "\"`),")
//	g.P("	),")
//	g.P("	fx.Annotate(")
//	g.P("		Register", clientName, "ProviderImpl,")
//	g.P("		fx.ResultTags(`name:\"", regName, "\"`),")
//	g.P("	),")
//	g.P("}}")
//
//	g.P()
//	g.P("func Register", clientName, "ProviderImpl(repo config.ConfigureWatcherRepo, logger log.Logger) (grpc.ClientConnInterface, error) {")
//	g.P("	server := &struct {")
//	g.P("		Server def.Server `conf_path:\"/registry/", regName, "/config\"`")
//	g.P("	}{")
//	g.P("		Server: def.Server{}, ")
//	g.P("	}")
//	g.P("	err := repo.LoadAndStart(server)")
//	g.P("	if err != nil {")
//	g.P("		return nil, err")
//	g.P("	}")
//	g.P("	return client.NewClientConn(&server.Server, logger)")
//	g.P("}")
//	g.P()
//}
