import './App.css';
import NewJobs from './components/NewJobs.js';
import OutputPanel from './components/OutputPanel.js';

import React, {Component} from 'react';


class App extends Component {

  constructor(props) {
    super(props);

    this.state = {
      apiserver: "localhost:8081"
    }

    this.handleAPIServerChange = this.handleAPIServerChange.bind(this);
  }

  handleAPIServerChange(e) {
    this.setState({
      apiserver: e.target.value,
    });
  }

  render() {
    return (
      <div className='App'>
        <header className='App-header'>
           Transaction App
        </header>
        <div className='App-body'>
          <div className='App-newJobs'>
            <NewJobs apiserver={this.state.apiserver}></NewJobs>
          </div>
          <div className='App-outputPanel'>
            <OutputPanel apiserver={this.state.apiserver} onChangeAPIServer={this.handleAPIServerChange}></OutputPanel>
          </div>
        </div>
      </div>
    );
  }
}

export default App;
