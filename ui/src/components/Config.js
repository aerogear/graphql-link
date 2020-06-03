import React from "react";
import {Spinner} from "@patternfly/react-core";

export const ConfigContext = React.createContext({config: null, onReloadConfig: null, onStoreConfig: null});

const Config = ({children}) => {
  let [config, setConfig] = React.useState(null);

  const onReloadConfig = () => {
    (async ()=>{
      try {
        let d = await fetch('/admin/config')
        d = await d.json()
        setConfig(d);
      } catch (error) {
        console.log(error)
      }
    })()
  };

  const onStoreConfig = async (config) => {
    try {
      await fetch('/admin/config', {
        method: 'post',
        body: JSON.stringify(config)
      })
      await onReloadConfig()
    } catch (error) {
      console.log(error)
    }
  };
  React.useEffect(onReloadConfig, []); // The empty array causes this effect to only run on mount
  if (config == null) {
    return <Spinner></Spinner>
  }
  return (
    <ConfigContext.Provider value={{config, onReloadConfig, onStoreConfig}}>{children}</ConfigContext.Provider>
  );
};


export default Config
