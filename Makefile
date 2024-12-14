VERSION=v0.0.0

start-chat: 
	@go run chatr.go chat
start-forwarder: 
	@go run chatr.go forwarder
start-match: 
	@go run chatr.go match
start-uploader: 
	@go run chatr.go uploader
start-user: 
	@go run chatr.go user
wire: 
	wire gen ./internal/wire 
proto-gen:
	protoc proto/*/*.proto --go_out=. --go-grpc_out=.
docker-compose-up:
	@docker-compose -f ./deployment/docker-compose.yaml up -d
docker-compose-down:
	@docker-compose -f ./deployment/docker-compose.yaml down
docker-compose-ps:
	@docker-compose -f ./deployment/docker-compose.yaml ps
docker-build:
	@docker build -f ./build/Dockerfile --build-arg VERSION=$(VERSION) -t thyyl/chatr:latest .