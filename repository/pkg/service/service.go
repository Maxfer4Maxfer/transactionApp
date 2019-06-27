package service

import (
	"context"
	"errors"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	stdopentracing "github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
	"github.com/google/uuid"

	"worker/gokit/workertransport"
	"repository/pkg/model"
)

// Service describes a service that represents repository.
type Service interface {
	RegisterNode(ctx context.Context, name string, IP string, port string) (string, error)
	GetAllNodes(ctx context.Context) ([]model.Node, error)
	NewJob(ctx context.Context) (string, error)
}

// Storage stores nodes
type Storage interface {
	NewNode(n model.Node) (model.NodeID, error)
	SaveNode(n model.Node) error
	GetAllNodes() ([]model.Node, error)
	DeleteNode(model.NodeID)
}

// New returns a basic Service with all of the expected middlewares wired in.
func New(s Storage, logger log.Logger, registerNodes, getAllNodes, newJobs metrics.Counter) Service {
	
	repo := Repo{s, logger}

	var svc Service
	{
		svc = LoggingMiddleware(logger)(repo)
		svc = InstrumentingMiddleware(registerNodes, getAllNodes, newJobs)(svc)
	}

	// Start checking nodes in a repository.
	checkNodesClose := make(chan struct{}, 1)
	go repo.CheckNodes(checkNodesClose)

	return svc
}

var (
	// ErrRepoUnevailable allows say that something wrong happens with a connection to DB
	ErrRepoUnevailable = errors.New("can't connect to a storage service")

	// ErrNodeAlreadyExist prevents users add a node with dublicate name
	ErrNodeAlreadyExist = errors.New("node with same name already registered in repo")

	// ErrEmptyRepo shows that a repo is empty
	ErrEmptyRepo = errors.New("empty repository")
)

// Repo implements Service interface
type Repo struct {
	s      Storage
	logger log.Logger
}


func (r Repo) RegisterNode(ctx context.Context, name string, IP string, port string) (string, error) {
	node := model.Node{
		Name: name,
		IP:   IP,
		Port: port,
	}

	id, err := r.s.NewNode(node)
	return id.String(), err
}

func (r Repo) GetAllNodes(ctx context.Context) ([]model.Node, error) {
	return r.s.GetAllNodes()
}

// FindFree returns a node with with low jobs running level
func (r Repo) FindFree(ctx context.Context) ( string, string, string, string, error) {
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
func (r Repo) NewJob(ctx context.Context) (string, error) {
	// id, name, IP, port, err := r.FindFree()
	id, name, IP, port, err := r.FindFree(ctx)
	r.logger.Log("method", "NewJob", "connecting to ", id+" "+name)

	grpcAddr := IP + port
	ctx, close := context.WithTimeout(ctx, time.Second)
	defer close()
	conn, err := grpc.DialContext(ctx, grpcAddr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		r.logger.Log("method", "NewJob", "err", err)
	}
	defer conn.Close()

	otTracer := stdopentracing.GlobalTracer() // no-op
	svc := workertransport.NewGRPCClient(conn, otTracer, r.logger)

	jID, err := svc.NewJob(ctx)
	r.logger.Log("method", "NewJob", "job ID", jID)

	return jID, err
}

// CheckNodes starts checking each node in the repository and save current number of running jobs
// checkNodesClose is a channel that should be close for stopping checking proccess.
func (r Repo) CheckNodes(checkNodesClose chan struct{}) error {
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

					jobsCount, err := svc.Ping(ctx)
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
					jobs, err := svc.GetJobs(ctx)
					if err != nil {
						r.logger.Log("method", "CheckNodes", "err", err)
					} else {
						r.logger.Log("method", "CheckNodes", "jobs", len(jobs))
						js := make([]model.Job, 0)
						for _, j := range jobs {
							id, _ := uuid.Parse(j.ID.String())
							job := model.Job{
								ID:         model.JobID{id},
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
