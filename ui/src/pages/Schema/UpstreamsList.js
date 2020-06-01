import {
  Button,
  DataList,
  DataListCell,
  DataListItem,
  DataListItemCells,
  DataListItemRow,
  DataToolbar,
  DataToolbarContent,
  DataToolbarItem,
  Flex,
  FlexItem,
  FlexModifiers
} from '@patternfly/react-core';

import React from 'react';
import {useMasterDetailState} from "../../components/HasDetails";


const UpstreamList = ({setDetails, details}) => {

  const {selected, onSelect} = useMasterDetailState(setDetails, details, "UpstreamDetails", id => ({id}));

  return (
    <React.Fragment>

      <DataToolbar id={"test"}>
        <DataToolbarContent>
          <DataToolbarItem>Upstream Servers</DataToolbarItem>
          <DataToolbarItem variant="separator"/>
          <DataToolbarItem><Button variant="primary">Add</Button></DataToolbarItem>
        </DataToolbarContent>
      </DataToolbar>

      <DataList aria-label="data list" selectedDataListItemId={selected}
                onSelectDataListItem={onSelect}>
        <DataListItem aria-labelledby="selectable-action-item1" id="content-padding-item1">
          <DataListItemRow>
            <DataListItemCells
              dataListCells={[
                <DataListCell key="primary content">
                  <Flex breakpointMods={[{modifier: FlexModifiers.column}]}>
                    <FlexItem>
                      <p><strong>anilist</strong></p>
                      <small>https://pf4.patternfly.org/</small>
                    </FlexItem>
                  </Flex>
                </DataListCell>,
              ]}
            />
          </DataListItemRow>
        </DataListItem>
      </DataList>
    </React.Fragment>
  )

};
export default UpstreamList