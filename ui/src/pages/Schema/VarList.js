import {
  Button,
  DataList,
  DataListCell,
  DataListItem,
  DataListItemCells,
  DataListItemRow, FormGroup,
  Split,
  SplitItem,
  TextContent,
  TextInput,
} from '@patternfly/react-core';
import React from 'react';
import {clone} from "../../utils";
import {TimesIcon} from "@patternfly/react-icons";

const VarList = ({value, onChange}) => {
  value = value || []

  const onAdd = () => {
    const c = clone(value)
    c.push({name: "$", value: ""})
    onChange(c)
  }

  const onNameChange = (key, newName) => {
    const c = clone(value)
    c[key].name = newName
    onChange(c)
  }

  const onValueChange = (key, newValue) => {
    const c = clone(value)
    c[key].value = newValue
    onChange(c)
  }

  const onDelete = (key) => {
    const c = clone(value)
    c.splice(key, 1)
    onChange(c)
  }

  console.log("value", value)
  return (
    <React.Fragment>
      <FormGroup fieldId="vars"
                 helperText="Additional variables to select from the current node to use in the upstream query">

        <div class="flex items-center mb-2">
          <div class="flex-grow left">
            <label className="pf-c-form__label" htmlFor="vars">
              <span className="pf-c-form__label-text">Variables</span>
            </label>
          </div>
          <div class="flex-none">
            <Button onClick={onAdd}>Add</Button>
          </div>
        </div>

        <DataList aria-label="data list">
          {value.map((item, key) =>
            <DataListItem id={"" + key} aria-labelledby={"" + key} key={key}>
              <DataListItemRow>
                <DataListItemCells
                  dataListCells={[
                    <DataListCell key="primary content">

                      <Button className="float-right" variant="plain" onClick={x => onDelete(key)}><TimesIcon/></Button>
                      <FormGroup label="Name" fieldId="name" isRequired>
                        <TextInput
                          id="field" name="field"
                          value={item.name} onChange={x => onNameChange(key, x)}
                          isRequired type="text"
                        />
                      </FormGroup>

                      <FormGroup label="Selection" fieldId="value" isRequired>
                        <TextInput
                          id="value" name="value"
                          value={item.value} onChange={x => onValueChange(key, x)}
                          isRequired type="text"
                        />
                      </FormGroup>


                    </DataListCell>,
                  ]}
                />
              </DataListItemRow>
            </DataListItem>
          )}
        </DataList>
      </FormGroup>
    </React.Fragment>
  )
};
export default VarList