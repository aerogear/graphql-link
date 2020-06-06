import React from 'react';
import {Redirect, Route, Switch, useRouteMatch} from "react-router-dom";
import TypeList from "./TypeList";
import Type from "./Type";
import Config, {ConfigContext} from "../../components/Config";

const Index = () => {
  let match = useRouteMatch();
  return (
    <Config>
      <ConfigContext.Consumer>
        {({config, onStoreConfig}) => (
          <Switch>
            <Route path={`${match.url}/types/:typeName`}>
              <Type config={config} onStoreConfig={onStoreConfig}/>
            </Route>
            <Route path={`${match.url}`}>
              <TypeList config={config} onStoreConfig={onStoreConfig}/>
            </Route>
            <Redirect to={`${match.url}`}/>
          </Switch>
        )}
      </ConfigContext.Consumer>
    </Config>
  )
}

export default Index
