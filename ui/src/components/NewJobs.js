import './NewJobs.css';
import React, {Component} from 'react';
import Form from 'react-bootstrap/Form';
import InputGroup from 'react-bootstrap/InputGroup';
import Button from 'react-bootstrap/Button';


class NewJobs extends Component {
  constructor(props) {
    super(props);
    this.state = {
      quantity: 5,
      interval: 3,
    } 

    this.handleChangeQuantity = this.handleChangeQuantity.bind(this);
    this.handleChangeInterval = this.handleChangeInterval.bind(this);
    this.newJob = this.newJob.bind(this);
    this.handleSubmit = this.handleSubmit.bind(this);
  }

  newJob() {
    fetch(
      'http://'+this.props.apiserver+'/newjob',
      {method: 'POST', body: JSON.stringify({})})
      .then(response => {
        return response.json();
      })
      .then(result => {console.log(result)});
  }

  handleChangeQuantity(event) {
    this.setState({
      quantity: event.target.value
    });
  }

  handleChangeInterval(event) {
    this.setState({
      interval: event.target.value
    });
  }


  sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
  }

  async handleSubmit(event) {
    event.preventDefault();
    for (var i = 0; i < this.state.quantity; i++) {
      this.newJob();
      await this.sleep(this.state.interval * 1000)
    }
  }


  render() {
    return (
      <div className='NewJobs-body'>
      <Form onSubmit={this.handleSubmit} className='NewJobs-inputForm'>
        <Form.Row>
          <InputGroup size="sm" className="NewJobs-inputGroup-Quantity">
            <InputGroup.Prepend>
              <InputGroup.Text id="inputGroup-Quantity">Quantity</InputGroup.Text>
            </InputGroup.Prepend>
            <Form.Control aria-label="Quantity" aria-describedby="inputGroup-Quantity" type='number' value={this.state.quantity} onChange={this.handleChangeQuantity} />
          </InputGroup>
          <InputGroup size="sm" className="NewJobs-inputGroup-Interval">
            <InputGroup.Prepend>
              <InputGroup.Text id="inputGroup-Interval">Interval</InputGroup.Text>
            </InputGroup.Prepend>
            <Form.Control aria-label="Interval" aria-describedby="inputGroup-Interval" type='number' value={this.state.interval} onChange={this.handleChangeInterval} />
          </InputGroup>
        </Form.Row>
        <Form.Row>
          <Button variant="outline-secondary" size="sm"  type='submit' block>Run Jobs</Button> 
        </Form.Row>
      </Form>
      </div>)
  }
}

export default NewJobs;
