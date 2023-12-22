package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path"
	"strings"
	"time"

	"github.com/thirdmartini/go-nvme"
	"github.com/thirdmartini/go-nvme/api"
	"github.com/thirdmartini/go-nvme/pkg/tracer"
	"github.com/thirdmartini/go-nvme/targets"

	"github.com/google/uuid"
)

var targetNum = uint64(0)

type TargetState struct {
	Id          string
	Name        string
	Description string
	Size        uint64
	Path        string
}

func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

func setStatus(w http.ResponseWriter, code int, status string) {
	stat := api.Status{
		Code:    code,
		Message: status,
	}
	w.WriteHeader(code)
	data, _ := json.Marshal(&stat)
	w.Write(data)

	fmt.Printf("SetStatus(%s)\n", status)
}

func receive(r *http.Request, req interface{}) error {
	defer r.Body.Close()

	data, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, req)
	if err != nil {
		return err
	}
	return nil
}

func respond(w http.ResponseWriter, resp interface{}) error {
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	w.Write(data)
	return nil
}

func main() {
	debugLevel := flag.Uint64("debug", 0, "Sets the debug level of the server")
	bindAddress := flag.String("bind", "", "Bind to a specific address")
	flag.Parse()

	if *debugLevel != 0 {
		tracer.AbortOnTodo = true
	}

	localIP := *bindAddress
	if localIP == "" {
		localIP = GetLocalIP()
	}

	addrFlag := localIP + ":4420"

	fmt.Printf("Local Address: %s\n", localIP)
	fmt.Printf("NVME Server: %s\n", addrFlag)
	go func() {
		profileAddress := localIP + ":6060"
		fmt.Printf("Profiler started at http://%s/debug/pprof\n", profileAddress)
		fmt.Printf("    go tool pprof -http :8080 'http://%s/debug/pprof/profile?seconds=10'\n", profileAddress)
		log.Println(http.ListenAndServe(profileAddress, nil))
	}()

	conf, err := LoadConfig("./targets.yaml")
	if err != nil {
		panic(err)
	}

	s, err := nvme.New(addrFlag)
	if err != nil {
		panic(err)
	}
	s.SetDebugLevel(*debugLevel)

	for _, t := range conf.Targets {
		id, err := uuid.Parse(t.UUID)
		if err != nil {
			fmt.Printf("Error: Target %s missing target UUID\n", t.Name)
			continue
		}

		target, err := targets.New(t.Type, t.Options)
		if err != nil {
			panic(err)
		}
		target.Start()

		subsys := &nvme.TargetSubsystem{
			NQN:             t.Name,
			Target:          target,
			ModelName:       t.ModelName,
			SerialNumber:    t.SerialNumber,
			FirmwareVersion: t.FirmwareVersion,
		}
		copy(subsys.UUID[:], id[:])

		fmt.Printf("Registering Target: %s (%s)\n", t.Name, t.Type)
		fmt.Printf("  Subsys: %+v\n", subsys)
		s.AddSubSystem(subsys)
	}

	files, err := os.ReadDir("./data")
	if err == nil {
		for _, file := range files {
			targetConfig := path.Join("./data", file.Name())
			if !strings.HasSuffix(targetConfig, ".json") {
				continue
			}

			fmt.Printf("Trying Target: %s\n", targetConfig)
			data, err := os.ReadFile(targetConfig)
			if err != nil {
				fmt.Printf("skip 2: %s %s\n", targetConfig, err.Error())
				continue
			}

			targetState := TargetState{}
			err = json.Unmarshal(data, &targetState)
			if err != nil {
				fmt.Printf("skip 3: %s %s  --> [%+v]\n", targetConfig, err.Error(), targetState)
				continue
			}

			id, err := uuid.Parse(targetState.Id)
			if err != nil {
				fmt.Printf("skip 4: %s %s\n", targetConfig, err.Error())
				continue
			}

			opt := targets.Options{
				"image": targetState.Path,
			}

			target, err := targets.New("file", opt)
			if err != nil {
				continue
			}
			target.Start()

			subsys := &nvme.TargetSubsystem{
				NQN:             fmt.Sprintf("nqn.2014-08.com.thirdmartini:uuid:%s", id.String()),
				Target:          target,
				ModelName:       "ThirdMartini NVME",
				SerialNumber:    id.String(),
				FirmwareVersion: "0.1.0",
			}
			copy(subsys.UUID[:], id[:])
			fmt.Printf("Registering Target: %s (%s) -> %+v\n", subsys.NQN, "file", subsys.UUID)
			fmt.Printf("  Subsys: %+v\n", subsys)
			s.AddSubSystem(subsys)
		}
	}
	fmt.Printf("WEB UI: http://%s:%s/\n", localIP, "8090")

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

		http.HandleFunc("/api/v1/CreateVolumeRequest", func(w http.ResponseWriter, r *http.Request) {
			log.Printf("Create Volume\n")
			req := api.CreateVolumeRequest{}
			resp := api.CreateVolumeResponse{}
			err := receive(r, &req)
			if err != nil {
				setStatus(w, http.StatusBadRequest, err.Error())
				return
			}

			id := uuid.New()
			raw := fmt.Sprintf("data/%s.raw", id.String())

			tmp, err := os.Create(raw)
			if err != nil {
				setStatus(w, http.StatusBadRequest, err.Error())
				return
			}
			tmp.Close()

			err = os.Truncate(raw, int64(req.Size))
			if err != nil {
				setStatus(w, http.StatusBadRequest, err.Error())
				return
			}

			opt := targets.Options{
				"image": raw,
			}

			targetState := TargetState{
				Id:          id.String(),
				Name:        req.Name,
				Description: req.Description,
				Size:        req.Size,
				Path:        raw,
			}

			state, err := json.Marshal(&targetState)
			if err != nil {
				setStatus(w, http.StatusBadRequest, err.Error())
				return

			}
			err = os.WriteFile(fmt.Sprintf("%s.json", raw), state, 0644)
			if err != nil {
				setStatus(w, http.StatusBadRequest, err.Error())
				return
			}

			target, err := targets.New("file", opt)
			if err != nil {
				setStatus(w, http.StatusBadRequest, err.Error())
				return
			}
			target.Start()

			subsys := &nvme.TargetSubsystem{
				NQN:             fmt.Sprintf("nqn.2014-08.com.thirdmartini:uuid:%s", id.String()),
				Target:          target,
				ModelName:       "ThirdMartini NVME",
				SerialNumber:    id.String(),
				FirmwareVersion: "0.1.0",
			}
			copy(subsys.UUID[:], id[:])
			fmt.Printf("Registering Target: %s (%s) -> %+v\n", subsys.NQN, "file", subsys.UUID)
			fmt.Printf("  Subsys: %+v\n", subsys)
			s.AddSubSystem(subsys)

			resp.Volume = &api.Volume{
				Name: id.String(),
				UUID: id.String(),
				Size: req.Size,
			}

			respond(w, &resp)
		})

		http.HandleFunc("/api/v1/DeleteVolumeRequest", func(w http.ResponseWriter, r *http.Request) {
			req := api.DeleteVolumeRequest{}
			resp := api.DeleteVolumeResponse{}

			err := receive(r, &req)
			if err != nil {
				setStatus(w, http.StatusBadRequest, err.Error())
				return
			}

			subSystems := s.ListSubSystems()
			for idx := range subSystems {
				subSys, ok := subSystems[idx].(*nvme.TargetSubsystem)
				if !ok {
					continue
				}

				uid, _ := uuid.FromBytes(subSys.UUID[:])
				if req.UUID == uid.String() {
					err = s.RemoveSubSystem(subSys.GetNQN())
					if err != nil {
						setStatus(w, http.StatusBadRequest, err.Error())
						return
					}

					target := subSys.Target.(*targets.FileTarget)
					target.Close()
					for _, kv := range target.GetRuntimeDetails() {
						if kv.Key == "image" {
							os.Remove(kv.Value)
						}
					}

					respond(w, &resp)
					return
				}
			}
			setStatus(w, http.StatusBadRequest, "does not exist")
		})

		http.HandleFunc("/api/v1/ListVolumeRequest", func(w http.ResponseWriter, r *http.Request) {
			fmt.Printf("")

			req := api.ListVolumeRequest{}
			resp := api.ListVolumeResponse{}

			err := receive(r, &req)
			if err != nil {
				setStatus(w, http.StatusBadRequest, "bad request")
				return
			}

			subSystems := s.ListSubSystems()
			for idx := range subSystems {
				subsys, ok := subSystems[idx].(*nvme.TargetSubsystem)
				if !ok {
					continue
				}

				uid, _ := uuid.FromBytes(subsys.UUID[:])
				volume := api.Volume{
					UUID: uid.String(),
					Name: uid.String(),
					Size: subsys.Target.GetSize(),
					NQN:  subsys.GetNQN(),
				}
				resp.Volumes = append(resp.Volumes, volume)
			}
			respond(w, &resp)
		})

		http.HandleFunc("/api/v1/GetVolumeRequest", func(w http.ResponseWriter, r *http.Request) {
			req := api.GetVolumeRequest{}
			resp := api.GetVolumeResponse{}

			err := receive(r, &req)
			if err != nil {
				setStatus(w, http.StatusBadRequest, "bad request")
				return
			}

			subSystems := s.ListSubSystems()
			for idx := range subSystems {
				subsys, ok := subSystems[idx].(*nvme.TargetSubsystem)
				if !ok {
					continue
				}

				uid, _ := uuid.FromBytes(subsys.UUID[:])
				if req.UUID == uid.String() {
					resp.Volume = &api.Volume{
						UUID: uid.String(),
						Name: uid.String(),
						Size: subsys.Target.GetSize(),
						NQN:  subsys.GetNQN(),
					}
					respond(w, &resp)
					return
				}
			}
			setStatus(w, http.StatusBadRequest, "does not exist")
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
	err = s.Serve()
	if err != nil {
		panic(err)
	}
}
