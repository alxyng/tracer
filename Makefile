GOBUILD=GOOS=linux GOARCH=arm64 go build
RPI_ADDR=pi@192.168.1.104
RPI_DIR=/home/pi/tracer

all: build
build:
	$(GOBUILD) -o build/api cmd/api/main.go
	$(GOBUILD) -o build/controller cmd/controller/main.go
	$(GOBUILD) -o build/writer cmd/writer/main.go
deploy-api:
	scp build/api $(RPI_ADDR):$(RPI_DIR)/api
deploy-controller:
	scp build/controller $(RPI_ADDR):$(RPI_DIR)/controller
deploy-writer:
	scp build/writer $(RPI_ADDR):$(RPI_DIR)/writer
deploy: deploy-api deploy-controller deploy-writer
enable:
	ssh $(RPI_ADDR) "sudo systemctl enable tracer.api.service"
	ssh $(RPI_ADDR) "sudo systemctl enable tracer.controller.service"
	ssh $(RPI_ADDR) "sudo systemctl enable tracer.writer.service"
disable:
	ssh $(RPI_ADDR) "sudo systemctl disable tracer.api.service"
	ssh $(RPI_ADDR) "sudo systemctl disable tracer.controller.service"
	ssh $(RPI_ADDR) "sudo systemctl disable tracer.writer.service"
start:
	ssh $(RPI_ADDR) "sudo systemctl start tracer.api.service"
	ssh $(RPI_ADDR) "sudo systemctl start tracer.controller.service"
	ssh $(RPI_ADDR) "sudo systemctl start tracer.writer.service"
stop:
	ssh $(RPI_ADDR) "sudo systemctl stop tracer.api.service"
	ssh $(RPI_ADDR) "sudo systemctl stop tracer.controller.service"
	ssh $(RPI_ADDR) "sudo systemctl stop tracer.writer.service"
status:
	ssh $(RPI_ADDR) "sudo systemctl status tracer.api.service"
	ssh $(RPI_ADDR) "sudo systemctl status tracer.controller.service"
	ssh $(RPI_ADDR) "sudo systemctl status tracer.writer.service"
cycle: build stop deploy start status

.PHONY: build deploy-api deploy-controller deploy-writer enable disable start stop status
