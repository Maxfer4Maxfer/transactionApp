package repo

import (
	"context"
	"errors"
	"time"

	"worker/gokit/workertransport"

	stdopentracing "github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"

	"github.com/go-kit/kit/log"
	"github.com/google/uuid"
)

// Repo implements Repo interface
type Repo struct {
	s      Storage
	logger log.Logger
}

// New create new repository of nodes which stored in object behind the Storage interface
func New(s Storage, logger log.Logger) Repo {
	r := Repo{s, logger}
	return r
}

var (
	// ErrRepoUnevailable allows say that something wrong happens with a connection to DB
	ErrRepoUnevailable = errors.New("can't connect to a storage service")

	// ErrNodeAlreadyExist prevents users add a node with dublicate name
	ErrNodeAlreadyExist = errors.New("node with same name already registered in repo")

	// ErrEmptyRepo shows that a repo is empty
	ErrEmptyRepo = errors.New("empty repository")
)

func (r *Repo) RegisterNode(name string, IP string, port string) (string, error) {
	// TODO: Validate input parameters

	node := Node{
		Name: name,
		IP:   IP,
		Port: port,
	}

	id, err := r.s.NewNode(node)
	return id.String(), err
}

func (r *Repo) GetAllNodes() ([]Node, error) {
	return r.s.GetAllNodes()
}

// FindFree returns a node with with low jobs running level
func (r *Repo) FindFree() (string, string, string, string, error) {
	nodes, err := r.s.GetAllNodes()
	if err != nil {
		return "", "", "", "", err
	}

	if len(nodes) == 0 {
		return "", "", "", "", ErrEmptyRepo
	}

	num := 0 //number of node with low jobs

	for k, n := range nodes {
		if n.JobsCount == 0 {
			return n.ID.String(), n.Name, n.IP, n.Port, nil
		}
		if n.JobsCount < nodes[num].JobsCount {
			num = k
		}
	}

	return nodes[num].ID.String(), nodes[num].Name, nodes[num].IP, nodes[num].Port, nil
}

// NewJob starts new job on a free node
func (r *Repo) NewJob() (string, error) {
	// id, name, IP, port, err := r.FindFree()
	id, name, IP, port, err := r.FindFree()
	r.logger.Log("method", "NewJob", "connecting to ", id+" "+name)

	grpcAddr := IP + port
	ctx, close := context.WithTimeout(context.Background(), time.Second)
	defer close()
	conn, err := grpc.DialContext(ctx, grpcAddr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		r.logger.Log("method", "NewJob", "err", err)
	}
	defer conn.Close()

	otTracer := stdopentracing.GlobalTracer() // no-op
	svc := workertransport.NewGRPCClient(conn, otTracer, r.logger)

	jID, err := svc.NewJob(context.Background())
	r.logger.Log("method", "NewJob", "job ID", jID)

	return jID, err
}

// CheckNodes starts checking each node in the repository and save current number of running jobs
// checkNodesClose is a channel that should be close for stopping checking proccess.
func (r *Repo) CheckNodes(checkNodesClose chan struct{}) error {
	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for range ticker.C {
			nodes, _ := r.s.GetAllNodes()
			for _, n := range nodes {
				// r.logger.Log("method", "CheckNodes", "node", n.ID.String()+" "+n.IP+n.Port)

				grpcAddr := n.IP + n.Port
				ctx, close := context.WithTimeout(context.Background(), time.Second)
				defer close()
				conn, err := grpc.DialContext(ctx, grpcAddr, grpc.WithInsecure(), grpc.WithBlock())
				if err != nil {
					r.logger.Log("method", "CheckNodes", "err", err)
					r.s.DeleteNode(n.ID)
					r.logger.Log("method", "CheckNodes", "action", "node deleted from the repository")
				} else {
					defer conn.Close()
					otTracer := stdopentracing.GlobalTracer() // no-op
					svc := workertransport.NewGRPCClient(conn, otTracer, r.logger)

					jobsCount, err := svc.Ping(context.Background())
					if err != nil {
						r.logger.Log("method", "CheckNodes", "err", err)
						r.s.DeleteNode(n.ID)
						r.logger.Log("method", "CheckNodes", "action", "node deleted from the repository")
					} else {
						// r.logger.Log("method", "CheckNodes", "jobsCount", jobsCount)
						// save a updated node
						n.JobsCount = jobsCount
						r.s.SaveNode(n)
					}
					jobs, err := svc.GetJobs(context.Background())
					if err != nil {
						r.logger.Log("method", "CheckNodes", "err", err)
					} else {
						r.logger.Log("method", "CheckNodes", "jobs", len(jobs))
						js := make([]Job, 0)
						for _, j := range jobs {
							id, _ := uuid.Parse(j.ID.String())
							job := Job{
								ID:         JobID{id},
								Per:        j.Per,
								Duration:   float32(j.Duration.Seconds()),
								StartTime:  j.StartTime,
								FinishTime: j.FinishTime,
							}
							js = append(js, job)
						}
						n.Jobs = js
						r.s.SaveNode(n)
					}
				}
			}
		}
	}()
	<-checkNodesClose
	ticker.Stop()
	return nil
}
