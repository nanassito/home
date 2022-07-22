set -eux


for target in switches ;
do
    protoc \
        -I . \
        --go_out=./proto \
        --go_opt=paths=source_relative \
        --go-grpc_out=./proto \
        --go-grpc_opt=paths=source_relative \
        --experimental_allow_proto3_optional \
        --grpc-gateway_out ./proto \
        --grpc-gateway_opt paths=source_relative \
        --grpc-gateway_opt generate_unbound_methods=true \
        ${target}/${target}.proto
done


for target in switches ;
do
    go build -o ./${target}/${target}.bin ./${target}/
done