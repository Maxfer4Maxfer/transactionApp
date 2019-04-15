package worker

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/nats-io/go-nats"
	"github.com/shirou/gopsutil/cpu"
)

const (
	tickerPeriod          = 1000 * time.Millisecond //update frequency
	idealTime             = 4 * time.Second         // idial time for perform one job
	idealMgz              = 4800                    // Mgz count for perform one job in IDEALTIME
	idealMgzForOnePercent = idealMgz / 100
)

// Worker implements Repo interface
type Worker struct {
	mtx      sync.RWMutex
	jobs     map[JobID]*Job
	stop     chan struct{}
	name     string
	IP       string
	port     string
	natsAddr string
	logger   log.Logger
	CPUCount int
	CPUModel string
	CPUCores int32
	tCPUMhz  int32
}

// New create new repository of nodes which stored in object behind the Storage interface
func New(name string, IP string, port string, natsAddr string, logger log.Logger) Worker {

	is, _ := cpu.Info()

	var tCPUMhz int32
	for _, c := range is {
		mn := c.ModelName
		fl, _ := strconv.ParseFloat(mn[len(mn)-7:len(mn)-3], 32)
		tCPUMhz = tCPUMhz + c.Cores*int32(fl*1000)
	}

	// for simplicity and not to bother yourself take tCPUMhz as idealMgz
	tCPUMhz = idealMgz

	logger.Log("worker", "New", "tCPUMhz", tCPUMhz)

	w := Worker{
		jobs:     make(map[JobID]*Job),
		stop:     make(chan struct{}),
		name:     name,
		IP:       IP,
		port:     port,
		natsAddr: natsAddr,
		logger:   logger,
		CPUCount: len(is),
		CPUModel: is[0].ModelName,
		CPUCores: is[0].Cores,
		tCPUMhz:  tCPUMhz,
	}

	logger.Log("worker", "New", "Worker", fmt.Sprint(w))

	// register new worker in a reposiroty.
	// NATS -> pb.Repo -> RegisterNode
	err := w.registerItself()
	if err != nil {
		logger.Log("worker", "New", "err", err)
		os.Exit(1)
	}
	go w.updateJobsStatus()

	return w
}

func (w *Worker) Stop() {
	w.stop <- struct{}{}
}

func (w *Worker) Ping() (int, error) {
	return w.activeJobsLen(), nil
}

func (w *Worker) NewJob() (ID JobID) {

	job := NewJob()

	w.mtx.Lock()
	defer w.mtx.Unlock()

	w.jobs[job.ID] = job

	return job.ID
}

func (w *Worker) GetJobs() []Job {
	w.mtx.RLock()
	defer w.mtx.RUnlock()
	jobs := make([]Job, 0)
	for _, j := range w.jobs {
		jobs = append(jobs, *j)
	}
	return jobs
}

func (w *Worker) updateJobsStatus() {
	ticker := time.NewTicker(tickerPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			for _, j := range w.jobs {
				if j.Per < 100 {
					mgzForJob := int(w.tCPUMhz) / w.activeJobsLen() // number of Mgz for performing current job
					j.Per = j.Per + (float32(tickerPeriod)/float32(idealTime))*(float32(mgzForJob)/float32(idealMgzForOnePercent))
					j.Duration = time.Since(j.StartTime)
					if j.Per >= 100 {
						j.Finish()
					}
				}
			}
		case <-w.stop:
			return
		}
	}
}

func (w *Worker) activeJobsLen() int {
	len := 0
	for _, j := range w.jobs {
		if j.Per != 100 {
			len = len + 1
		}
	}
	return len
}

func (w *Worker) registerItself() error {

	var waitG sync.WaitGroup
	waitG.Add(1)

	ticker := time.NewTicker(1 * time.Second)

	go func() {
		for range ticker.C {
			// tryint to connect to NATS
			nc, err := nats.Connect(w.natsAddr)
			if err != nil {
				w.logger.Log("method", "registerItself", "natsAddr", w.natsAddr, "err", err)
			} else {
				// prepering the message to a RegisterNode method
				str := `{"name":"` + w.name + `", "ip":"` + w.IP + `", "port":"` + w.port + `"}`
				// tryint to call a method RegisterNode
				r, err := nc.Request("RegisterNode", []byte(str), 10*time.Second)
				if err != nil {
					w.logger.Log("method", "registerItself", "action", "RegisterNode call", "err", err)
				} else {
					// creating a structure for parsing a responce
					var resp struct {
						String string `json:"str"`
						Err    string `json:"err"`
					}
					// encoding the response
					switch err = json.Unmarshal(r.Data, &resp); {
					case err != nil:
						w.logger.Log("method", "registerItself", "action", "Unmarshal a RegisterNode responce", "err", err)
					case resp.Err != "":
						w.logger.Log("method", "registerItself", "action", "Something wrong on the repository site", "err", resp.Err)
					default:
						w.logger.Log("method", "registerItself", "message", "Node registered succesfully", "id", string(r.Data))
						ticker.Stop()
						waitG.Done()
					}
				}
			}
		}
	}()

	waitG.Wait()
	return nil
}
