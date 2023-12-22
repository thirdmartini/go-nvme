package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"sync/atomic"
	"time"

	uuid2 "github.com/google/uuid"
	"github.com/thirdmartini/go-nvme"
	"github.com/thirdmartini/go-nvme/client"
	"github.com/thirdmartini/go-nvme/pkg/tracer"
	"github.com/thirdmartini/go-nvme/targets"
)

func main() {
	kind := flag.String("kind", "write", "")
	queueCount := flag.Uint("qc", 1, "")
	queueDepth := flag.Uint("qd", 1, "")
	debug := flag.Bool("debug", false, "")

	serverAddress := flag.String("server", "localhost:4444", "")
	webAddress := flag.String("web", "localhost:8080", "")
	profileAddress := flag.String("prof", "localhost:6060", "")
	flag.Parse()

	go func() {
		fmt.Printf("Profiler started at http://%s/debug/pprof\n", *profileAddress)
		fmt.Printf("    go tool pprof go tool pprof -http :8080 'http://%s/debug/pprof/profile?seconds=10'\n", *profileAddress)
		log.Println(http.ListenAndServe(*profileAddress, nil))
	}()

	s, err := nvme.New(*serverAddress)
	if err != nil {
		log.Fatal(err)
	}

	if *debug {
		s.SetDebugLevel(tracer.TraceAll)
	} else {
		s.SetDebugLevel(0)
	}

	/*
		nullTarget, err := targets.defaultFactory.New("null", nil)
		if err != nil {
			panic(err)
		}*/

	nullTarget, err := targets.New("sleepy", nil)
	if err != nil {
		panic(err)
	}
	nullTarget.Start()

	subsys := &nvme.TargetSubsystem{
		NQN:    "nqn.2020-20.com.thirdmartini.nvme:null",
		Target: nullTarget,
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

	fmt.Printf("WEB UI: http://%s/\n", *webAddress)

	go func() {
		http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
			fmt.Fprintf(w, "<html><pre>\n")
			fmt.Fprintf(w, "<h2><a href=\"/sessions\">Sessions</a> | <a href=\"/targets\">Targets</a> </h2><hr>\n")
			fmt.Fprintf(w, "</pre></html>\n")

		})

		http.HandleFunc("/sessions", func(w http.ResponseWriter, req *http.Request) {
			fmt.Fprintf(w, "<html><pre>\n")
			fmt.Fprintf(w, "<h2>Sessions</a> | <a href=\"/targets\">Targets</a></h2><hr>\n")
			sessions := s.GetSessions()

			for idx := range sessions {
				s := sessions[idx]
				fmt.Fprintf(w, "%s : %s\n", s.Source, s.Ctrl.ConnectedHostNQN)
				fmt.Fprintf(w, "  %s\n", s.Ctrl.ConnectedSubNQN)
				fmt.Fprintf(w, "       Controller: %d       QueueID: %d\n", s.Ctrl.ControllerID, s.Ctrl.QueueID)
				fmt.Fprintf(w, "       Queue Head: %d     QueueSize: %d\n", s.Ctrl.SQHD, s.Ctrl.QueueSize)
				fmt.Fprintf(w, "     Request Count: %d   RequestTime: (%s/request)\n", s.Ctrl.RequestCount, s.Ctrl.RequestTime/time.Duration(s.Ctrl.RequestCount))
			}
			fmt.Fprintf(w, "</pre></html>\n")
		})

		http.HandleFunc("/targets", func(w http.ResponseWriter, req *http.Request) {
			fmt.Fprintf(w, "<html><pre>\n")
			fmt.Fprintf(w, "<h2><a href=\"/sessions\">Sessions</a> | Targets</h2><hr>\n")
			subSystems := s.ListSubSystems()
			for idx := range subSystems {
				s := subSystems[idx]
				fmt.Fprintf(w, " %s\n", s.GetNQN())
				for _, v := range s.GetRuntimeDetails() {
					fmt.Fprintf(w, "    + %s: %s\n", v.Key, v.Value)
				}
			}
			fmt.Fprintf(w, "</pre></html>\n")
		})

		http.ListenAndServe(":8090", nil)
	}()

	tc, err := client.New(*serverAddress, "nqn.2020-20.com.thirdmartini.nvme:null")
	if err != nil {
		log.Fatal(err)
	}

	/*
		tc.WithTracer(&nvme.LogTracer{
			Id: "tc",
			Level: nvme.TraceAll,
		})*/

	status := tc.Login()
	if status.IsError() {
		panic(status.AsError())
	}

	io, status := tc.OpenIOQueue(1)
	if status.IsError() {
		panic(status.AsError())
	}

	data := make([]byte, 4096, 4096)

	count := 0
	tick := time.Now()
	tickMemStats := runtime.MemStats{}
	mem := runtime.MemStats{}
	runtime.ReadMemStats(&tickMemStats)

	// Cleanup stuff
	defer func() {
		tc.Close()
		s.Close()
	}()

	switch *kind {
	case "read":
		// Initial - 17777/sec
		// Base    - 19336/sec   A:735957 F:724025
		// Best    - 20495/sec   A:616109 F:619863
		// Perf    - 23200/sec   A:69618 F:69587
		// Perf:   - 23880/sec   A:1 F:0 ( All print statements commented out )
		// -- a few allocations come from the code that does this checking
		for {
			err = io.Read(1, data)
			if err != nil {
				panic(err)
			}
			count++

			if time.Now().Sub(tick) > time.Second {
				runtime.ReadMemStats(&mem)

				fmt.Printf("Perf: %d/sec   A:%d F:%d\n", count,
					mem.Mallocs-tickMemStats.Mallocs,
					mem.Frees-tickMemStats.Frees)
				count = 0

				tickMemStats = mem
				tick = time.Now()
			}
		}

	case "write":
		for {
			// Perf: 28719/sec   A:1 F:0
			err = io.Write(1, data)
			if err != nil {
				panic(err)
			}
			count++
			if time.Now().Sub(tick) > time.Second*1 {
				runtime.ReadMemStats(&mem)

				fmt.Printf("Perf: %d/sec   A:%d F:%d\n", count,
					mem.Mallocs-tickMemStats.Mallocs,
					mem.Frees-tickMemStats.Frees)
				count = 0

				tickMemStats = mem
				tick = time.Now()
			}
		}

	case "mread":
		count := uint64(0)

		for q := 0; q < int(*queueCount); q++ {
			iol, status := tc.OpenIOQueue(uint16(q + 1))
			if status.IsError() {
				panic(status.AsError())
			}

			for i := 0; i < int(*queueDepth); i++ {
				go func(iol *client.IOQueue) {
					ldata := make([]byte, 512, 512)
					for {
						err = iol.Read(1, ldata)
						if err != nil {
							panic(err)
						}
						atomic.AddUint64(&count, 1)
					}
				}(iol)
			}
		}

		ticker := time.NewTicker(time.Second)
		oldC := atomic.LoadUint64(&count)
		for {
			select {
			case <-ticker.C:
				runtime.ReadMemStats(&mem)

				newC := atomic.LoadUint64(&count)

				fmt.Printf("Perf: %d/sec   A:%d F:%d\n",
					newC-oldC,
					mem.Mallocs-tickMemStats.Mallocs,
					mem.Frees-tickMemStats.Frees)
				oldC = newC
				tickMemStats = mem
				tick = time.Now()
			}
		}

	case "mwrite":
		count := uint64(0)

		for q := 0; q < int(*queueCount); q++ {
			iol, status := tc.OpenIOQueue(uint16(q + 1))
			if status.IsError() {
				panic(status.AsError())
			}

			for i := 0; i < int(*queueDepth); i++ {
				go func(iol *client.IOQueue, lba uint64) {
					ldata := make([]byte, 512, 512)
					for {
						err = iol.Write(lba, ldata)
						if err != nil {
							panic(err)
						}
						atomic.AddUint64(&count, 1)
					}
				}(iol, 0)
			}
		}
		ticker := time.NewTicker(time.Second)

		oldC := atomic.LoadUint64(&count)
		for {
			select {
			case <-ticker.C:
				runtime.ReadMemStats(&mem)

				newC := atomic.LoadUint64(&count)

				fmt.Printf("Perf: %d/sec   A:%d F:%d\n",
					newC-oldC,
					mem.Mallocs-tickMemStats.Mallocs,
					mem.Frees-tickMemStats.Frees)
				oldC = newC
				tickMemStats = mem
				tick = time.Now()
			}
		}
	}

}
