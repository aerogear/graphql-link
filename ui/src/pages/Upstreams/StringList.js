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

const NameValueList = ({value, onChange, fieldId, label, helperText}) => {
  value = value || []

  const onAdd = () => {
    const c = clone(value)
    c.push("")
    onChange(c)
  }

  const onValueChange = (i, newValue) => {
    const c = clone(value)
    c[i] = newValue
    onChange(c)
  }

  const onDelete = (i) => {
    const c = clone(value)
    c.splice(i, 1)
    onChange(c)
  }

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
          {value.map((item, i) =>
            <DataListItem id={"" + i} aria-labelledby={"" + i} key={i}>
              <DataListItemRow>
                <DataListItemCells
                  dataListCells={[
                    <DataListCell key="primary content">
                      <Button className="float-right" variant="plain" onClick={x => onDelete(i)}><TimesIcon/></Button>
                      <TextInput
                        value={item} onChange={x => onValueChange(i, x)}
                        type="text"
                      />
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