import {
  Button,
  DataList,
  DataListCell,
  DataListItem,
  DataListItemCells,
  DataListItemRow,
  Divider,
  PageSection,
  PageSectionVariants,
  Split,
  SplitItem,
  TextContent
} from '@patternfly/react-core';
import React from 'react';
import {Link} from "react-router-dom";
import {clone} from "../../utils";

const TypeList = ({config, onStoreConfig}) => {

  const value = config.types

  const onAdd = () => {
    const c = clone(config)
    c.types.push({
      name: "Unnamed",
      actions: []
    })
    onStoreConfig(c)
  }

  return (
    <React.Fragment>
      <PageSection variant={PageSectionVariants.light}>
        <Split gutter="md">
          <SplitItem>
            <TextContent>
              <h1>Types</h1>
            </TextContent>
          </SplitItem>
          <SplitItem isFilled></SplitItem>
          <SplitItem><Button variant="primary" onClick={onAdd}>Add</Button></SplitItem>
        </Split>
      </PageSection>
      <Divider component="div"/>

      <PageSection noPadding={false}>

        <DataList aria-label="data list">
          {value.map((item, key) =>
            <DataListItem id={"" + key} aria-labelledby={"" + key} key={key}>
              <DataListItemRow>
                <DataListItemCells
                  dataListCells={[
                    <DataListCell key="primary content">
                      <Link to={`./types/${item.name}`}>
                        {item.name}
                      </Link>
                    </DataListCell>,
                  ]}
                />
              </DataListItemRow>
            </DataListItem>
          )}
        </DataList>

      </PageSection>

      {/*<UpstreamsList/>*/}
    </React.Fragment>
  )
};
export default TypeList