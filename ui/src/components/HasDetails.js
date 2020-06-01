import {Drawer, DrawerContent, DrawerContentBody} from '@patternfly/react-core';
import React from 'react';


export function useMasterDetailState(setDetails, details, detailsComponentName, propsFn) {
  const [selected, setSelected] = React.useState(null);
  const onClose = () => {
    setSelected(null)
    setDetails(null)
  }

  const onSelect = (id) => {
    setSelected(id)
    setDetails({
      component: detailsComponentName,
      props: {onClose: onClose, ...propsFn(id)}
    })
  }
  React.useEffect(() => {
    if (details == null || details.component == detailsComponentName) {
      setSelected(null)
    }
  }, [details]);
  return {selected, onSelect};
}

const HasDetails = ({children, details, components}) => {

  let panelContent = <div></div>;


  if (details != null) {
    const {component, props = {}} = details
    const Component = components[component]
    panelContent = <Component {...props}></Component>
  }

  return (
    <Drawer isExpanded={details != null}>
      <DrawerContent panelContent={panelContent} className={'pf-m-no-background'}>
        <DrawerContentBody hasPadding>{children}</DrawerContentBody>
      </DrawerContent>
    </Drawer>
  )
};
export default HasDetails