import {
  Button,
  DataList,
  DataListCell,
  DataListItem,
  DataListItemCells,
  DataListItemRow,
  Flex,
  FlexItem,
  FlexModifiers,
  Split,
  SplitItem,
  TextContent,
} from '@patternfly/react-core';
import React from 'react';
import DetailsPanel, {DetailsClose} from "../../components/DetailsPanel";
import ActionDetails from "./ActionDetails";
import {clone} from "../../utils";

const ActionList = ({config, actions, setActions}) => {
  actions = actions || []
  const [selected, setSelected] = React.useState(null);

  const onAdd = () => {
    const c = clone(actions)
    c.push({
      type: "mount",
    })
    setActions(c)
  }

  return (
    <React.Fragment>

      <div className={["mb-4"]}>
        <Split gutter="md">
          <SplitItem>
            <TextContent>
              <h1>Actions</h1>
            </TextContent>
          </SplitItem>
          <SplitItem isFilled></SplitItem>
          <SplitItem><Button variant="primary" onClick={onAdd}>Add</Button></SplitItem>
        </Split>
      </div>


      <DataList aria-label="data list" selectedDataListItemId={selected} onSelectDataListItem={setSelected}>
        {actions.map((item, key) =>
          <DataListItem id={"" + key} aria-labelledby={"" + key} key={key}>
            {
              selected === ("" + key) &&
              <DetailsPanel id="right-panel">
                <ActionDetails
                  config={config}
                  onClose={_ => setSelected(null)}
                  id={key}
                  actions={actions}
                  setActions={setActions}
                ></ActionDetails>
                <DetailsClose onClick={_ => setSelected(null)}></DetailsClose>
              </DetailsPanel>
            }
            <DataListItemRow>
              <DataListItemCells
                dataListCells={[
                  <DataListCell key="primary content">
                    <Flex breakpointMods={[{modifier: FlexModifiers.column}]}>
                      <FlexItem>
                        <p><strong>#{key + 1}: {item.type}</strong></p>
                        {
                          item.type === "mount" && item.field === "" &&
                          <small><b>All fields from</b> {item.upstream} <b>/</b> {item.query}</small>
                        }
                        {
                          item.type === "mount" && item.field !== "" &&
                          <small>{item.field} <b>=</b>  {item.upstream} <b>/</b> {item.query}</small>
                        }
                        {
                          item.type === "link" &&
                          <small>{item.field} <b>=</b> {item.upstream} <b>/</b> {item.query}</small>
                        }
                        {
                          item.type === "remove" &&
                          <small className="line-through">{item.field}</small>
                        }
                        {
                          item.type === "rename" &&
                          <small><span className="line-through">{item.field}</span> <b>to</b> {item.to}</small>
                        }
                      </FlexItem>
                    </Flex>
                  </DataListCell>,
                ]}
              />
            </DataListItemRow>
          </DataListItem>
        )}
      </DataList>

    </React.Fragment>
  )
};
export default ActionList