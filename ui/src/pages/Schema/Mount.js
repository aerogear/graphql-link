import {FormGroup, SelectOption, TextArea, TextInput} from '@patternfly/react-core';

import React from 'react';
import BetterSelect from "../../components/BetterSelect";
import {fieldSetters} from "../../utils";

const Mount = ({upstreams, action, setAction}) => {

  const onChange = fieldSetters(Object.assign({
    field: "",
    upstream: "",
    query: ""
  }, action), setAction)

  return <React.Fragment>
    <FormGroup label="Field" fieldId="field"
               helperText="The new field name that will be created on the gateway schema. If not set, then all fields of the upstream query will be mounted.">
      <TextInput
        id="field" name="field"
        value={action.field} onChange={onChange.field}
        type="text"
      />
    </FormGroup>

    <FormGroup label="Upstream Server" isRequired fieldId="upstream" helperText="The upstream sever to send the request to">
      <BetterSelect value={action.upstream} setValue={onChange.upstream}>
        {upstreams.map((item, key) =>
          <SelectOption value={item.name} key={key}>{item.name}</SelectOption>
        )}
      </BetterSelect>
    </FormGroup>

    <FormGroup label="Path Query" isRequired fieldId="query" helperText="A path query defines which upstream graph node gets mounted in the gateway">
      <TextArea
        id="query" name="query"
        value={action.query} onChange={onChange.query}
        isRequired type="text"
      />
    </FormGroup>
  </React.Fragment>
}

export default Mount