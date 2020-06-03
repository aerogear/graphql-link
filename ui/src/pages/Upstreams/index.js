import React from 'react';
import Config, {ConfigContext} from "../../components/Config";
import Page from "./Page";

export default () => {
  return (
    <Config>
      <ConfigContext.Consumer>
        {({config, onStoreConfig}) => (
          <Page config={config} onStoreConfig={onStoreConfig}></Page>
        )}
      </ConfigContext.Consumer>
    </Config>
  )
}