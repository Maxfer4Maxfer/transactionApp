import './WorkerList.css';
import React, {Component} from 'react';
import JobList from './JobList.js';
import Form from 'react-bootstrap/Form';



class WorkerList extends Component {
    constructor(props) {
        super(props);
 
        this.state = {
            lastN: true,
        } 

        this.handleInputChange = this.handleInputChange.bind(this);
        this.renderWorker = this.renderWorker.bind(this);
    }

    handleInputChange(event) {
        const target = event.target;
        const value = target.type === 'checkbox' ? target.checked : target.value;
        const name = target.name;
    
        this.setState({
          [name]: value
        });
      }

    renderJobs(jobs){
        if (jobs.length === 0) {
            return this.renderMessage("There is no jobs for this worker.")
        } else {
            return (
                <JobList jobs={jobs} errorMessage=""></JobList>
            )
        }
    }
    renderWorker(w) {
        var jobs = w.jobs
        if (this.state.lastN) {
            jobs = w.jobs.slice(0,5)
        }
        return(
            <div className='WorkerList-worker'>
                <table className='WorkerList-table'>
                    <thead className='WorkerList-thead'>
                    <tr>
                        <th>Worker ID</th>
                        <th>Name</th> 
                        <th>IP</th>
                    </tr>
                    </thead>
                    <tr className='WorkerList-table-worker'>
                        <th>{w.id}</th>
                        <th>{w.name}</th> 
                        <th>{w.ip}</th>
                    </tr>
                </table>
                {this.renderJobs(jobs)}
            </div>
        )
    }

    renderError() {
        return (
            <div className='WorkerList-Error'>
                {this.props.errorMessage}
                <br />
                Please go to the Settings tab and change a API Server address.
            </div>
        )
    }
    
    renderMessage(message) {
        return (
            <div className='WorkerList-Message'>
                {message}
            </div>
        )
    }

    renderWorkers() {
        if (this.props.workers.length === 0) {
            return this.renderMessage("There is no workers registered in the repository.")
        } else {
            return (
                <div className="WorkerList">
                <Form className="WorkerList-jobs-count">
                    <Form.Check
                        inline
                        name="lastN"
                        type="checkbox"
                        checked={this.state.lastN}
                        onChange={this.handleInputChange} 
                        label="Show last 5 jobs for each worker"
                    />
                </Form>
                    {this.props.workers.map((w) => this.renderWorker(w))}
                </div>
            )
        }
    }

    
    render() {
        if (this.props.errorMessage !== "") {
            return this.renderError()
        } else {
            return this.renderWorkers()
        }
    }
}

export default WorkerList;
