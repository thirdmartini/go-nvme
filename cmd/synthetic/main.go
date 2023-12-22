package main

import (
	"bytes"
	"fmt"
	"log"

	uuid2 "github.com/google/uuid"
	"github.com/thirdmartini/go-nvme"
	"github.com/thirdmartini/go-nvme/client"
	"github.com/thirdmartini/go-nvme/pkg/tracer"
	"github.com/thirdmartini/go-nvme/targets"
)

func fillBuffer(buffer []byte, pattern byte) {
	for ofs := 0; ofs < len(buffer); ofs++ {
		buffer[ofs] = pattern
	}
}

func main() {
	log.Println("whats up here!")

	s, err := nvme.New("localhost:4444")
	if err != nil {
		log.Fatal(err)
	}

	s.SetDebugLevel(tracer.TraceAll)

	target := &targets.MemTarget{
		Buffer: make([]byte, 1024*512, 1024*512),
	}
	target.Start()

	for i := 0; i < 1024; i++ {
		fillBuffer(target.Buffer[i*512:i*512+512], byte(i))
	}

	subsys := &nvme.TargetSubsystem{
		NQN:    "nqn.2020-20.com.thirdmartini.nvme:null",
		Target: target,
	}
	uuid, err := uuid2.NewUUID()
	if err != nil {
		log.Fatal(err)
	}
	copy(subsys.UUID[:], uuid[:])

	s.AddSubSystem(subsys)

	go func() {
		err = s.Serve()
		fmt.Printf("Server Exited\n")
		if err != nil {
			panic(err)
		}
	}()

	client, err := client.New("localhost:4444", "nqn.2020-20.com.thirdmartini.nvme:null")
	if err != nil {
		log.Fatal(err)
	}

	client.WithTracer(&tracer.LogTracer{
		Id:    "client",
		Level: tracer.TraceAll,
	})

	status := client.Login()
	if status.IsError() {
		panic(status.AsError())
	}

	/*
		_, err = client.AdminQueue().GetProperty(0, 1)
		if err != nil {
			panic(err)
		}
		if err != nil {
			log.Fatal(err)
		}

		_, err = client.AdminQueue().GetProperty(0, 1)
		if err != nil {
			panic(err)
		}

		client.AdminQueue().KeepAlive()*/

	io, status := client.OpenIOQueue(1)
	if status.IsError() {
		panic(status.AsError())
	}

	data := make([]byte, 512, 512)
	err = io.Read(1, data)
	if err != nil {
		panic(err)
	}

	compare := make([]byte, 512, 512)
	fillBuffer(compare, byte(1))
	if bytes.Compare(data, compare) != 0 {
		panic("bad compare")
	}

	err = io.Read(100, data)
	if err != nil {
		panic(err)
	}

	/*
		data = make([]byte, 512*1024, 512*1024)
		err = io.Read(0, data)
		if err != nil {
			panic(err)
		}
		if bytes.Compare(data, target.Buffer) != 0 {
			panic("bad compare")
		}*/

	fillBuffer(data, 1)
	err = io.Write(0, data)
	if err != nil {
		panic(err)
	}

	client.Close()
	fmt.Printf("Client.Close() -- done\n")
	s.Close()
	fmt.Printf("Server.Close() -- done\n")
}
