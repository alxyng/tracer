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
start-api:
	ssh $(RPI_ADDR) "sudo systemctl start tracer.api.service"
start-controller:
	ssh $(RPI_ADDR) "sudo systemctl start tracer.controller.service"
start-writer:
	ssh $(RPI_ADDR) "sudo systemctl start tracer.writer.service"
start: start-api start-controller start-writer
stop-api:
	ssh $(RPI_ADDR) "sudo systemctl stop tracer.api.service"
stop-controller:
	ssh $(RPI_ADDR) "sudo systemctl stop tracer.controller.service"
stop-writer:
	ssh $(RPI_ADDR) "sudo systemctl stop tracer.writer.service"
stop: stop-api stop-controller stop-writer
status-api:
	ssh $(RPI_ADDR) "sudo systemctl status tracer.api.service"
status-controller:
	ssh $(RPI_ADDR) "sudo systemctl status tracer.controller.service"
status-writer:
	ssh $(RPI_ADDR) "sudo systemctl status tracer.writer.service"
status: status-api status-controller status-writer
cycle-api: build stop-api deploy-api start-api status-api
cycle-controller: build stop-controller deploy-controller start-controller status-controller
cycle-writer: build stop-writer deploy-writer start-writer status-writer
cycle: cycle-api cycle-controller cycle-writer

.PHONY: build enable disable
.PHONY: deploy-api deploy-controller deploy-writer
.PHONY: start-api start-controller start-writer
.PHONY: stop-api stop-controller stop-writer
.PHONY: status-api status-controller status-writer
.PHONY: cycle-api cycle-controller cycle-writer
