import {
  Button,
  DataList,
  DataListCell,
  DataListItem,
  DataListItemCells,
  DataListItemRow,
  Divider, EmptyState, EmptyStateIcon,
  PageSection,
  PageSectionVariants,
  Split,
  SplitItem,
  TextContent, Title
} from '@patternfly/react-core';
import React from 'react';
import {Link} from "react-router-dom";
import {clone} from "../../utils";
import {SearchIcon} from "@patternfly/react-icons";

const TypeList = ({config, onStoreConfig}) => {

  const value = config.types || []

  const onAdd = () => {
    const c = clone(config)
    c.types = c.types || []
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

      <PageSection noPadding={false}>
        {value.length === 0 && <EmptyState>
          No types defined yet.
        </EmptyState>
        }
        {value.length !== 0 &&
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
        }

      </PageSection>

    </React.Fragment>
  )
};
export default TypeList