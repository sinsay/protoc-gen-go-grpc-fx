# protoc-gen-go-grpc-fx

create from protoc-gen-go-grpc, add feature for kratos with fx injection framework,

usage:

```
  protoc --go-grpc-fx_out=. --go-grpc-fx_opt=require_unimplemented_servers=false[,other options...] \
```

## Feature added

- add injection method for fx
- won't break the original code
