import * as React from 'react';
import {Route, Switch} from 'react-router-dom';
import {LastLocationProvider} from 'react-router-last-location';

import {NotFound} from '../pages/NotFound';
import routes from '../routes';
import {useDocumentTitle} from '../utils';


const PageNotFound = ({title}) => {
  useDocumentTitle(title);
  return <Route component={NotFound}/>;
};

export default () => {

  return <LastLocationProvider>
    <Switch>
      {routes.map(({path, exact, component, title, isAsync}, idx) => (
        <Route path={path}  exact={exact}  component={component} key={idx}/>
      ))}
      <PageNotFound title="404 Page Not Found"/>
    </Switch>
  </LastLocationProvider>
}