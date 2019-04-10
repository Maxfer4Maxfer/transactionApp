import './JobList.css';
import React, {Component} from 'react';

class JobList extends Component {
    constructor(props) {
        super(props);
 
        this.state = {
        } 
    }

    renderError() {
        return (
            <div className='JobList-Error'>
                {this.props.errorMessage}
                <br />
                Please go to the Settings tab and change a API Server address.
            </div>
        )
    }
    
    renderMessage(message) {
        let nMessage = message.split ('\n').map ((item, i) => <div key={i}>{item}</div>);
        return (
            <div className='JobList-Message'>
                {nMessage}
            </div>
        )
    }

    renderJob(j) {
        return(
            <tr className='JobList-job'>
                <th>{j.id}</th>
                <th>{j.worker}</th> 
                <th>{j.duration}</th>
                <th>{j.startTime}</th>
                <th>{j.finishTime}</th>
                <th>{j.state}</th>
            </tr>
        )
    }

    renderJobs() {
        if (this.props.jobs.length === 0) {
            return this.renderMessage('There are no jobs started.\nRun a new job and it immediately appeared in this list')
        } else {
            var jobs = this.props.jobs
            return (
                <div className="JobList">
                <table className="JobList-table">
                    <thead className="JobList-thead">
                        <tr>
                            <th>ID</th>
                            <th>Worker</th> 
                            <th>Duration</th>
                            <th>Start Time</th>
                            <th>Finish Time</th>
                            <th>✓</th>
                        </tr>
                    </thead>
                    {jobs.map((j) => this.renderJob(j))}
                </table>    
                </div>
            )
        }
    }

    render() {
        if (this.props.errorMessage !== "") {
            return this.renderError()
        } else {
            return this.renderJobs()
        }
    }
}

export default JobList;
