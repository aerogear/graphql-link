import {FormGroup, SelectOption, TextInput} from '@patternfly/react-core';

import React from 'react';
import BetterSelect from "../../components/BetterSelect";
import {fieldSetters} from "../../utils";

const Rename = ({action, setAction}) => {

  const onChange = fieldSetters(Object.assign({
    field: "",
    to: ""
  }, action), setAction)

  return <React.Fragment>
    <FormGroup label="Field" fieldId="field" isRequired
               helperText="The field name to be renamed">
      <TextInput
        id="field" name="field"
        value={action.field} onChange={onChange.field}
        isRequired type="text"
      />
    </FormGroup>

    <FormGroup label="To" fieldId="to" isRequired
               helperText="The the new name to give the field">
      <TextInput
        id="to" name="to"
        value={action.to} onChange={onChange.to}
        isRequired type="text"
      />
    </FormGroup>

  </React.Fragment>
}

export default Rename