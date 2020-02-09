################################################################################
## @Author:					Thomas Bouder <Tbouder>
## @Email:					Tbouder@protonmail.com
## @Date:					Sunday 05 January 2020 - 19:54:37
## @Filename:				makefile
##
## @Last modified by:		Tbouder
## @Last modified time:		Sunday 05 January 2020 - 19:55:03
################################################################################

SERVICE=Pictures
SERVICE_PB=${SERVICE}.pb.go
SERVICE_PROTO=${SERVICE}.proto
SERVICE_PROTO_ALT=Albums.proto
SERVICE_PACKAGE=pictures

all: init proto build

init:
	@-echo "Creating sdk directory"
	@-mkdir -p ../../sdk && mkdir -p ../../sdk/${SERVICE}

proto:
	@-make init
	@-echo "Generating Proto file"
	@-protoc --go_out=plugins=grpc,import_path=main:./ ${SERVICE_PROTO}
	@-protoc --go_out=plugins=grpc,import_path=main:./ ${SERVICE_PROTO_ALT}
	@-protoc --go_out=plugins=grpc,import_path=${SERVICE_PACKAGE}:../../sdk/${SERVICE} ${SERVICE_PROTO}
	@-protoc --go_out=plugins=grpc,import_path=${SERVICE_PACKAGE}:../../sdk/${SERVICE} ${SERVICE_PROTO_ALT}

build:
	docker build -t piwigo__grpc__${SERVICE_PACKAGE} .

clean:
	rm -rf ${SERVICE_PB}
	rm -rf .env
	# rm -rf wait-for-it.sh
