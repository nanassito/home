set -eux


GOOS="linux"
GOARCH="amd64"
for app in netscan app mqtt_json_2_str mqtt_crash_corrector;
do
    go build -o ./bin/${app} ./cmd/${app}/${app}.go || go build -o ./bin/${app} ./cmd/${app}/main.go
done