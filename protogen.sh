#!/bin/sh

DAPR_PROTO_ROOT=https://raw.githubusercontent.com/dapr/dapr/master/dapr/proto/

go install github.com/gogo/protobuf/gogoreplace

echo "Download proto files from dapr/dapr."
wget -q ${DAPR_PROTO_ROOT}/common/v1/common.proto -O ./dapr/proto/common/v1/common.proto
gogoreplace 'option go_package = "github.com/dapr/dapr/pkg/proto/common/v1;common";' 'option go_package = "github.com/dapr/go-sdk/dapr/proto/common/v1;common";' ./dapr/proto/common/v1/common.proto

wget -q ${DAPR_PROTO_ROOT}/runtime/v1/appcallback.proto -O ./dapr/proto/runtime/v1/appcallback.proto
gogoreplace 'option go_package = "github.com/dapr/dapr/pkg/proto/runtime/v1;runtime";' 'option go_package = "github.com/dapr/go-sdk/dapr/proto/runtime/v1;runtime";' ./dapr/proto/runtime/v1/appcallback.proto

wget -q ${DAPR_PROTO_ROOT}/runtime/v1/dapr.proto -O ./dapr/proto/runtime/v1/dapr.proto
gogoreplace 'option go_package = "github.com/dapr/dapr/pkg/proto/runtime/v1;runtime";' 'option go_package = "github.com/dapr/go-sdk/dapr/proto/runtime/v1;runtime";' ./dapr/proto/runtime/v1/dapr.proto

echo "Generating gRPC Proto clients"
protoc -I . ./dapr/proto/common/v1/*.proto --go_out=plugins=grpc:../../../
protoc -I . ./dapr/proto/runtime/v1/*.proto --go_out=plugins=grpc:../../../

echo "Clean up proto files"
rm -f ./dapr/proto/common/v1/*.proto
rm -f ./dapr/proto/runtime/v1/*.proto
