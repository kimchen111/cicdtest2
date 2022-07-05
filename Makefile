all: build

build:
	mkdir -p bin
	swag init
	go build -o bin/controller main.go
	go build -ldflags '-linkmode "external" -extldflags "-static"' -o bin/vpeagent vpe/main.go
	go build -ldflags '-linkmode "external" -extldflags "-static"' -o bin/cpeagent cpe/main.go

pack:
	tar -czf sdwan.tar.gz bin tests

run-ctrl:
	swag init
	go run main.go

#run-vpe:
#	go run vpe/main.go

#run-cpe:
#	go run cpe/main.go

#rcpe:
#	go build -ldflags '-linkmode "external" -extldflags "-static"' -o bin/cpeagent cpe/main.go
#	ssh root@192.168.3.1 "/root/stopca"
#	scp bin/cpeagent root@192.168.3.1:/root
#	ssh root@192.168.3.1 "/root/cpeagent 172.26.213.127"

#rvpe:
#	go build -o bin/vpeagent vpe/main.go
#	ssh root@192.168.3.254 "/root/stopva"
#	scp bin/vpeagent root@192.168.3.254:/root
#	ssh root@192.168.3.254 "/root/vpeagent 172.26.213.127"

clean:
	rm -f bin/*
