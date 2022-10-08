set -eux


for target in switches ;
do
    protoc \
        -I ./proto \
        --go_out=./pkg/${target}/proto \
        --go_opt=paths=source_relative \
        --go-grpc_out=./pkg/${target}/proto \
        --go-grpc_opt=paths=source_relative \
        --experimental_allow_proto3_optional \
        --grpc-gateway_out ./pkg/${target}/proto \
        --grpc-gateway_opt paths=source_relative \
        --grpc-gateway_opt generate_unbound_methods=true \
        proto/${target}.proto
done


GOOS="linux"
GOARCH="amd64"
for target in switches netscan;
do
    go build -o ./bin/${target} ./cmd/${target}/${target}.go
done