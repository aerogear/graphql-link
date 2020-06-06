import {FormGroup, TextInput} from '@patternfly/react-core';

import React from 'react';
import {fieldSetters} from "../../utils";

const Remove = ({action, setAction}) => {

  const onChange = fieldSetters(Object.assign({
    field: "",
  }, action), setAction)

  return <React.Fragment>
    <FormGroup label="Field" fieldId="field" isRequired
               helperText="The field name to be removed from the gateway schema">
      <TextInput
        id="field" name="field"
        value={action.field} onChange={onChange.field}
        isRequired type="text"
      />
    </FormGroup>

  </React.Fragment>
}

export default Remove