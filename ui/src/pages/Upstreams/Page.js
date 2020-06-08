import {
  Button,
  DataList,
  DataListCell,
  DataListItem,
  DataListItemCells,
  DataListItemRow, EmptyState,
  Flex,
  FlexItem,
  FlexModifiers,
  PageSection,
  PageSectionVariants,
  Split,
  SplitItem,
  TextContent
} from '@patternfly/react-core';
import React from 'react';
import {clone, fromKeyedArray, toKeyedArray} from "../../utils";
import DetailsPanel, {DetailsClose} from "../../components/DetailsPanel";
import Details from "./Details";

export default ({config, onStoreConfig}) => {

  const [selected, setSelected] = React.useState(null);
  const value = toKeyedArray(config.upstreams)
  const setValue = v => {
    const c = clone(config)
    c.upstreams = fromKeyedArray(v)
    onStoreConfig(c)
  }
  const add = () => {
    const c = clone(config)
    c.upstreams = c.upstreams || {}
    c.upstreams["Unnamed"] = {}
    onStoreConfig(c)

  }

  const onDetailsClose = () => {
    setSelected(null)
  }

  return <React.Fragment>
    <PageSection variant={PageSectionVariants.light}>
      <Split gutter="md">
        <SplitItem>
          <TextContent>
            <h1>Upstream Servers</h1>
          </TextContent>
        </SplitItem>
        <SplitItem isFilled></SplitItem>
        <SplitItem><Button variant="primary" onClick={add}>Add</Button></SplitItem>
      </Split>
    </PageSection>

    <PageSection>

      {value.length === 0 && <EmptyState>
        No upstream servers defined yet.
      </EmptyState>
      }
      {value.length !== 0 &&
      <DataList aria-label="data list" selectedDataListItemId={selected} onSelectDataListItem={setSelected}>
        {value.map((item, key) =>
          <DataListItem id={"" + key} aria-labelledby={"" + key} key={key}>
            {
              selected === ("" + key) &&
              <DetailsPanel id="right-panel">
                <Details id={parseInt(selected)} value={value} setValue={setValue}
                         onClose={onDetailsClose}></Details>
                <DetailsClose onClick={onDetailsClose}></DetailsClose>
              </DetailsPanel>
            }
            <DataListItemRow>
              <DataListItemCells
                dataListCells={[
                  <DataListCell key="primary content">
                    <Flex breakpointMods={[{modifier: FlexModifiers.column}]}>
                      <FlexItem>
                        <p><strong>{item.name}</strong></p>
                        <small>{item.url}</small>
                      </FlexItem>
                    </Flex>
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

}