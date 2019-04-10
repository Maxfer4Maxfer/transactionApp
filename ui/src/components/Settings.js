import './Settings.css';
import React, {Component} from 'react';
import Form from 'react-bootstrap/Form';
import InputGroup from 'react-bootstrap/InputGroup';
import Button from 'react-bootstrap/Button';
 

class Settings extends Component {
    constructor(props) {
        super(props);
 
        this.state = {
            lastN: true,
        } 

    }

    renderError() {
        return (
            <div className='Settings-Error'>
                {this.props.errorMessage}
            </div>
        )
    }
    render() {
      return (
        <div className='Settings-body'>
            <Form onSubmit={this.handleSubmit} className='Settings-inputForm'>
            <Form.Row>
                <InputGroup size="sm" className="Settings-inputGroup-APIServer">
                <InputGroup.Prepend>
                    <InputGroup.Text id="inputGroup-APIServer">API Server</InputGroup.Text>
                </InputGroup.Prepend>
                <Form.Control aria-label="APIServer" aria-describedby="inputGroup-APIServer" type='text' value={this.props.apiserver} onChange={this.props.onChangeAPIServer} />
                </InputGroup>
            </Form.Row>  
            </Form>
            {this.renderError()}
        </div>)
    }
}

export default Settings;
