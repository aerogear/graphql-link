import {
  Button,
  DataList,
  DataListCell,
  DataListItem,
  DataListItemCells,
  DataListItemRow,
  FormGroup,
  TextInput,
} from '@patternfly/react-core';
import React from 'react';
import {clone} from "../../utils";
import {TimesIcon} from "@patternfly/react-icons";

const NameValueList = ({value, onChange, fieldId, label, helperText, nameLabel, valueLabel}) => {
  value = value || []

  const onAdd = () => {
    const c = clone(value)
    c.push({name: "", value: ""})
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
      <FormGroup fieldId={fieldId}
                 helperText={helperText}>

        <div class="flex items-center mb-2">
          <div class="flex-grow left">
            <label className="pf-c-form__label" htmlFor={fieldId}>
              <span className="pf-c-form__label-text">{label}</span>
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
                      <FormGroup label={nameLabel} fieldId="name" isRequired>
                        <TextInput
                          id="field" name="field"
                          value={item.name} onChange={x => onNameChange(key, x)}
                          isRequired type="text"
                        />
                      </FormGroup>

                      <FormGroup label={valueLabel} fieldId="value">
                        <TextInput
                          id="value" name="value"
                          value={item.value} onChange={x => onValueChange(key, x)}
                          type="text"
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
export default NameValueList