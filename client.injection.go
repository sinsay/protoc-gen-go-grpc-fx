package main

import "google.golang.org/protobuf/compiler/protogen"

func GenerateClientInjection(clientName string, g *protogen.GeneratedFile, service *protogen.Service, gen *protogen.Plugin) {

	g.P("func New", clientName, " (cc ", grpcPackage.Ident("ClientConnInterface"), ") ", clientName, " {")
	helper.generateNewClientDefinitions(g, service, clientName)
	g.P("}")
	g.P()

	// Generate Service Name
	// try to get service name
	g.P("func (c *", unexport(service.GoName), "Client) RegisterNameForDiscover() string {")
	regName := getRegistryName(gen, service)
	g.P("return \"", regName, "\"")
	g.P("}")
	g.P()

	// generate init function
	g.P("func init() {")
	g.P("injection.InjectMany(Register", clientName, "Provider()...)")
	g.P("}")

	// generate method for create instance
	g.P("func Register", clientName, "Provider() []interface{} {")
	g.P("return []interface{}{")
	g.P("	fx.Annotate(")
	g.P("		New", clientName, ",")
	g.P("		fx.As(new(", clientName, ")),")
	g.P("		fx.ParamTags(`name:\"", regName, "\"`),")
	g.P("	),")
	g.P("	fx.Annotate(")
	g.P("		Register", clientName, "ProviderImpl,")
	g.P("		fx.ResultTags(`name:\"", regName, "\"`),")
	g.P("	),")
	g.P("}}")

	g.P()
	g.P("func Register", clientName, "ProviderImpl(repo config.ConfigureWatcherRepo, logger log.Logger) (grpc.ClientConnInterface, error) {")
	g.P("	server := &struct {")
	g.P("		Server def.Server `conf_path:\"/registry/", regName, "/config\"`")
	g.P("	}{")
	g.P("		Server: def.Server{}, ")
	g.P("	}")
	g.P("	err := repo.LoadAndStart(server)")
	g.P("	if err != nil {")
	g.P("		return nil, err")
	g.P("	}")
	g.P("	return client.NewClientConn(&server.Server, logger)")
	g.P("}")
	g.P()
}
