import {
  DrawerActions,
  DrawerCloseButton,
  DrawerHead,
  DrawerPanelBody,
  DrawerPanelContent,
  Flex,
  FlexItem,
  FlexModifiers,
  Title
} from '@patternfly/react-core';

import React from 'react';

const UpstreamDetails = ({onClose}) => {

  return (<DrawerPanelContent>
    <DrawerHead>
      <Title headingLevel="h2" size="xl">Details</Title>
      <DrawerActions>
        <DrawerCloseButton onClick={onClose}/>
      </DrawerActions>
    </DrawerHead>
    <DrawerPanelBody>
      <Flex breakpointMods={[{modifier: FlexModifiers["space-items-lg"]}, {modifier: FlexModifiers["column"]}]}>
        <FlexItem>
          <p>The content of the drawer really is up to you. It could have form fields, definition lists, text lists,
            labels, charts, progress bars, etc. Spacing recommendation is 24px margins. You can put tabs in here,
            and can also make the drawer scrollable.</p>
        </FlexItem>
      </Flex>
    </DrawerPanelBody>
  </DrawerPanelContent>)
};
export default UpstreamDetails