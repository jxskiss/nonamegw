.PHONY: gen_protocol

list:
	@$(MAKE) -pRrq -f $(lastword $(MAKEFILE_LIST)) : 2>/dev/null | awk -v RS= -F: '/^# File/,/^# Finished Make data base/ {if ($$1 !~ "^[#.]") {print $$1}}' | sort | egrep -v -e '^[^[:alnum:]]' -e '^$@$$'

gen_proto:
	cd proto && protoc --go_out=${GOPATH}/src packet.proto protocol.proto messag.proto data.proto brokersvc.proto cometsvc.proto bizapi.proto
	cd proto && protoc --go-grpc_out=./brokersvc --go-grpc_opt=paths=source_relative brokersvc.proto
	cd proto && protoc --go-grpc_out=./bizapi --go-grpc_opt=paths=source_relative bizapi.proto
