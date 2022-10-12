set -eux


for app in switches air ;
do
    target_dir="./pkg/${app}_proto"
    mkdir -p ${target_dir}
    protoc \
        -I ./proto \
        --go_out=${target_dir} \
        --go_opt=paths=source_relative \
        --go-grpc_out=${target_dir} \
        --go-grpc_opt=paths=source_relative \
        --experimental_allow_proto3_optional \
        --grpc-gateway_out ${target_dir} \
        --grpc-gateway_opt paths=source_relative \
        --grpc-gateway_opt generate_unbound_methods=true \
        proto/${app}.proto
done


GOOS="linux"
GOARCH="amd64"
for app in switches netscan air;
do
    go build -o ./bin/${app} ./cmd/${app}/${app}.go
done