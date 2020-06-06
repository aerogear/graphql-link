import React from 'react';
import ReactDOM from 'react-dom';

import App from './App/index';
import * as serviceWorker from './serviceWorker';

import './assets/main.css'

if ( window.self === window.top) {
  ReactDOM.render(
    <React.StrictMode>
      <App/>
    </React.StrictMode>,
    document.getElementById('root')
  );
} else {
  console.log("in an iframe... likely due to running in dev mode, and not proxying correctly to graphiql.  will do a manual redirect hack")
  window.location = "http://127.0.0.1:8080/graphiql"
}


// If you want your app to work offline and load faster, you can change
// unregister() to register() below. Note this comes with some pitfalls.
// Learn more about service workers: https://bit.ly/CRA-PWA
serviceWorker.unregister();
