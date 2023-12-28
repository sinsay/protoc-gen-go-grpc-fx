# protoc-gen-go-grpc-fx

create from protoc-gen-go-grpc, add feature for kratos with fx injection framework,

usage:

```
  protoc --go-grpc-fx_out=. --go-grpc-fx_opt=require_unimplemented_servers=false[,other options...] \
```

## Feature added

- add injection method for fx
- won't break the original code
- new generate function provide the basic ability to build  auto inject framework



## Usage

### Define API

first you need an protocol file like

```protobuf
syntax = "proto3";

package api.ping.service.pingservicev1;

// define the output package
option go_package = "github.com/you/package_name;pingservicev1";

import "google/api/annotations.proto";
import "google/protobuf/descriptor.proto";

// define service extend for naming service
extend google.protobuf.ServiceOptions {
  optional string name = 99999;
}


// PingReq is the demo ping api's argument
message PingReq {
  string message = 1;
}

// PingResp is the demo ping api's response
message PingResp {
  int32 code = 1;
  string message = 3;
}

// SrvPingV1 demo
service SrvPingV1 {

  // define the service name
  option (name) = "/ping_service/v1";

  // Ping api
  rpc Ping(pingv1.PingReq) returns (pingv1.PingResp);
}
```



now you can generate the output file in golang, the generated file keep the same ability as `protoc-gen-go-grpc` but has little more things as next:



### Server Side

on the server side, we generate new infomation for auto start server, the generate code like this:



```go

// Generate Injection
type RegisterSrvPingV1ServerGRPCResult struct{}

func (*RegisterSrvPingV1ServerGRPCResult) String() string {
	return "SrvPingV1ServerGRPCServer"
}

func RegisterSrvPingV1ServerGRPCProvider(newer interface{}) []interface{} {
	return []interface{}{
		// For provide dependency
		fx.Annotate(
			newer,
			fx.As(new(SrvPingV1Server)),
		),
		// For create instance
		fx.Annotate(
			RegisterSrvPingV1ServerProviderImpl,
			fx.As(new(fmt.Stringer)),
			fx.ResultTags(`group:"grpc_register"`),
		),
	}
}

// RegisterSrvPingV1ServerProviderImpl use to trigger register
func RegisterSrvPingV1ServerProviderImpl(s grpc.ServiceRegistrar, srv SrvPingV1Server) *RegisterSrvPingV1ServerGRPCResult {
	RegisterSrvPingV1Server(s, srv)
	return &RegisterSrvPingV1ServerGRPCResult{}
}
```



With this codes, we can load the grpc server with `fx`  injection framework like this

```go

// ping is the implementation of SrvPingV1
type ping struct {
	pingservicev1.UnimplementedSrvPingV1Server
}

// implementation
func (p *ping) Ping(ctx context.Context, in *pingservicev1.PingReq) (out *pingservicev1.PingResp, error) { 
  /* do something */ 
}

// The ping provider
func NewPing() *SrvPingV1 {
  return &ping{}
}

// the run can trigger the register action we generate
type injectDep struct{}

func run() {
  // the register of grpc server
  serverProvider := fx.pingservicev1.RegisterSrvPingV1ServerHTTPProvider(NewPing)
  
  fx.Provide(
    serverProvider,
    // the collector to trigger all server already reagister to fx
    fx.Annotate(
	func (grpcRegister []fmt.Stringer) *injectDep {
        	fmt.Print(len(grpcRegister))
		return &injectDep{}
	},
	fx.ParamTags(`name:"logger"`, `group:"http_register"`, `group:"grpc_register"`),
    ),
    // call while fx run
    fx.Invoke(
	func (n *injectDep) {
	    fmt.Print("Injected service finished...")
	},
    ),
  )
}

```




### Client Side

you can get service name defined on portocol file,

first is the client side, you got service name, and register provider for injection framework with `fx`

```go
type SrvPingV1Client interface {
	// Ping the demo api
	Ping(ctx context.Context, in *resources.PingReq, opts ...grpc.CallOption) (*resources.PingResp, error)
        // RegisterNameForDiscover is the service name for registry and service discover
	RegisterNameForDiscover() string
}

// ...

func (c *srvPingV1Client) RegisterNameForDiscover() string {
	return "/ping_service/v1"
}

func registerSrvPingV1ClientGRPCNameProvider() []string {
	return []string{"/ping_service/v1", "grpc"}
}

func RegisterSrvPingV1ClientGRPCProvider(creator interface{}) []interface{} {
	return []interface{}{
		fx.Annotate(
			NewSrvPingV1Client,
			fx.As(new(SrvPingV1Client)),
			fx.ParamTags(`name:"/ping_service/v1/grpc"`),
		),
		fx.Annotate(
			creator,
			fx.As(new(grpc.ClientConnInterface)),
			fx.ParamTags(`name:"/ping_service/v1/grpc/name"`),
			fx.ResultTags(`name:"/ping_service/v1/grpc"`),
		),
		fx.Annotate(
			registerSrvPingV1ClientGRPCNameProvider,
			fx.ResultTags(`name:"/ping_service/v1/grpc/name"`),
		),
	}
}

```

what is this use for? you can use it to create an injection with adding new dependence on the generated code,

on the example you can see an argument named `creator` on function `RegisterSrvPingV1ClientGRPCProvider`, it's the bridge between biz code and generated code, now you can do things on your personal biz code or develop framework like this:



```go
// configMap has like { "/ping_service/v1": "http://localhost:100001" }

// RegisterGRPCClient create and grpcConnInterface from service name/type
func RegisterGRPCClient(nameAndType []string) (grpc.ClientConnInterface, error) {
	if len(nameAndType) != 2 {
		return nil, errors.New("generated code for provide service name with wrong format, we need array like ['service_name', 'grpc']")
	}

	serviceName := nameAndType[0]
	serviceType := nameAndType[1]

	// load your client path from config center or config file etc, as example we define
  serverHost := configMap[serviceName]
  conn := newGRPCConn(serverHost)
  return conn, nil
}

// run run's the program with injection of grpc client
func run() {
  injection := pingservicev1.RegisterSrvPingV1ClientGRPCProvider(RegisterGRPCClient) 	
  invoke := fx.Invoke(func(client pingservicev1.SrvPingV1Client){
    // now the biz code can use grpc client without any initialize
    // client.Ping(context.Background(), &pingservicev1.PingReq{Message: "hello world"})
  })
  fx.New([]fx.Option{injection, invoke}).Run()
}

```



