package gorm

import (
	"time"
	
	_ "github.com/go-sql-driver/mysql"

	"github.com/go-kit/kit/log"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	
	repo "repository/pkg/model"
	"repository/pkg/service"
)

// Node represents an executer instance machine
type Node struct {
	ID        string `gorm:"primary_key"`
	Name      string
	IP        string
	Port      string
	JobsCount int
	Jobs      []Job `gorm:"foreignkey:NodeID"`
}

type Job struct {
	ID         string `gorm:"primary_key"`
	Per        float32
	Duration   float32
	StartTime  time.Time
	FinishTime time.Time
	NodeID     string
}

// NodeStorage implements mySQL storage for nodes
type NodeStorage struct {
	DB     *gorm.DB
	DSN    string
	logger log.Logger
}

// New create in memory repository for storing nodes
func New(dsn string, logger log.Logger) (*NodeStorage, func()) {

	ns := &NodeStorage{
		DB:     nil,
		DSN:    dsn,
		logger: logger,
	}

	closeCh := make(chan struct{}, 1)
	go ns.connectToDB(closeCh)
	closeFn := func() {
		closeCh <- struct{}{}
	}

	return ns, closeFn
}

func (ns *NodeStorage) connectToDB(closeCh chan struct{}) {
	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for range ticker.C {
			if ns.DB == nil {
				db, err := gorm.Open("mysql", ns.DSN)
				if err != nil {
					ns.logger.Log("db", "mysql", "message", "got an error", "err", err)
				} else {
					ns.logger.Log("db", "mysql", "message", "connection is established")
					ns.DB = db
					ns.DB.AutoMigrate(&Node{})
					ns.DB.AutoMigrate(&Job{})
				}

			}
			if ns.DB != nil {
				err := ns.DB.DB().Ping()
				if err != nil {
					ns.logger.Log("db", "mysql", "message", "lost connection to the db", "err", err)
				}
			}
		}
		ns.DB.Close()
	}()
	<-closeCh
	ticker.Stop()
}

func (ns *NodeStorage) NewNode(n repo.Node) (repo.NodeID, error) {
	if ns.DB != nil {

		id := uuid.New()
		node := Node{
			ID:        id.String(),
			Name:      n.Name,
			IP:        n.IP,
			Port:      n.Port,
			JobsCount: 0,
		}

		ns.DB.Create(&node)
		return repo.NodeID{id}, nil
	} else {
		return repo.NodeID{}, service.ErrRepoUnevailable
	}
}
func (ns *NodeStorage) SaveNode(n repo.Node) error {

	node := Node{
		ID:        n.ID.String(),
		Name:      n.Name,
		IP:        n.IP,
		Port:      n.Port,
		JobsCount: n.JobsCount,
		Jobs:      []Job{},
	}

	for _, j := range n.Jobs {
		node.Jobs = append(node.Jobs, Job{
			ID:         j.ID.String(),
			Per:        j.Per,
			Duration:   j.Duration,
			StartTime:  j.StartTime,
			FinishTime: j.FinishTime,
		})
	}

	ns.DB.Save(&node)

	return nil
}

func (ns *NodeStorage) GetAllNodes() ([]repo.Node, error) {

	nodes := []Node{}
	result := []repo.Node{}

	// err := ns.DB.DB().Ping()
	// if err == nil {
	if ns.DB != nil {
		if err := ns.DB.DB().Ping(); err == nil {
			ns.DB.Preload("Jobs").Find(&nodes)
			for _, n := range nodes {
				id, _ := uuid.Parse(n.ID)

				jobs := []repo.Job{}
				for _, j := range n.Jobs {
					id, _ := uuid.Parse(j.ID)
					jobs = append(jobs, repo.Job{
						ID:         repo.JobID{id},
						Per:        j.Per,
						Duration:   j.Duration,
						StartTime:  j.StartTime,
						FinishTime: j.FinishTime,
					})
				}

				result = append(result, repo.Node{
					ID:        repo.NodeID{id},
					Name:      n.Name,
					IP:        n.IP,
					Port:      n.Port,
					JobsCount: n.JobsCount,
					Jobs:      jobs,
				})
			}
		}
	}

	return result, nil
}

func (ns *NodeStorage) DeleteNode(id repo.NodeID) {
	node := Node{
		ID: id.String(),
	}

	ns.DB.Preload("Jobs").Find(&node)
	for _, j := range node.Jobs {
		ns.DB.Delete(&j)
	}

	ns.DB.Delete(&node)
}
