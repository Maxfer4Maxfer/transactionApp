import './OutputPanel.css';
import React, {Component} from 'react';
import Tabs from 'react-bootstrap/Tabs';
import Tab from 'react-bootstrap/Tab';
import JobList from './JobList.js';
import WorkerList from './WorkerList.js';
import Settings from './Settings.js';

class OutputPanel extends Component {
    constructor(props) {
        super(props);
        this.state = {
            alljobs: [],
            workers: [],
            errorMessage: "",
        } 
    }

    componentDidMount() {
        this.interval = setInterval(() => this.fetchingData(), 1000);
    }

    componentWillUnmount() {
        clearInterval(this.interval);
    }

    fetchingData() {
        fetch(
        'http://'+this.props.apiserver+'/getallnodes',
        {method: 'POST', body: JSON.stringify({})})
        .then(response => {
            if (!response.ok) {
                console.log("response.statusText");
                console.log(response.statusText);
                return null;
            } else {
                return response.json().then(result => {
                    var jobs = []
                    var workers = []
                    result.nodes.map(node => {
                        var worker = {
                            id: node.id, 
                            name: node.name,
                            ip: node.ip + ':' + node.port,
                            jobscount: node.jobscount,
                            jobs: []
                        }
                        node.jobs.map(j => {
                            var state = '⟳'
                            if (j.finishTime.substring(0, 1) !== '0') {
                                state = '✓'
                            }
                            var finishTime = ""
                            if (j.finishTime.substring(0,1) !== "0") {
                                finishTime = j.finishTime.substring(0,10) + ' ' + j.finishTime.substring(11,19)
                            }
                            var job = {
                                id: j.id.substring(0,8),
                                worker: node.name,
                                duration: j.duration.toFixed(2),
                                startTime: j.startTime.substring(0,10) + ' ' + j.startTime.substring(11,19),
                                finishTime: finishTime,
                                state: state,
                            }
                            jobs.push(job)
                            worker.jobs.push(job)
                        })
                        worker.jobs.sort((a, b) => (a.startTime < b.startTime) ? 1 : -1)
                        workers.push(worker)
                    })
                    
                    jobs.sort((a, b) => (a.startTime < b.startTime) ? 1 : -1)
                    this.setState({alljobs: jobs});

                    result.nodes.sort((a, b) => (a.id < b.id) ? 1 : -1)
                    this.setState({workers: workers});
                    
                    this.setState({errorMessage: ""});
                })
            }
        })
        .catch(error => {
            this.setState({
                alljobs: [],
                workers: [],
                errorMessage: error.message,
            });
        });
    }

    render() {
      return (
        <Tabs defaultActiveKey="allJobs" className='OutputPanel-tabs'>
            <Tab className='OutputPanel-tab' eventKey="allJobs" title="All Jobs">
                <JobList jobs={this.state.alljobs} errorMessage={this.state.errorMessage}></JobList>
            </Tab>
            <Tab className='OutputPanel-tab' eventKey="jobsByWorker" title="Jobs by Workers">
                <WorkerList workers={this.state.workers} errorMessage={this.state.errorMessage}></WorkerList>
            </Tab>
            <Tab className='OutputPanel-tab' eventKey="settings" title="Settings">
                <Settings apiserver={this.props.apiserver} errorMessage={this.state.errorMessage} onChangeAPIServer={this.props.onChangeAPIServer}></Settings>
            </Tab>
        </Tabs>
      )
    }
}

export default OutputPanel;
