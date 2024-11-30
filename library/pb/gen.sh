# shellcheck disable=SC2035
protoc -I ./ --go_out=. --go-grpc_out=. *.proto
protoc -I . --gotag_out=auto="json-as-camel+form-as-camel":. *.proto
protoc-go-inject-tag -input="*.pb.go"
