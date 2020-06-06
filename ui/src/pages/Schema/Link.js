import {FormGroup, SelectOption, TextArea, TextInput} from '@patternfly/react-core';

import React from 'react';
import BetterSelect from "../../components/BetterSelect";
import {fieldSetters} from "../../utils";
import VarList from "./VarList";

const Link = ({upstreams, action, setAction}) => {

  const onChange = fieldSetters(Object.assign({
    field: "",
    upstream: "",
    query: "",
    vars: []
  }, action), setAction)

  return <React.Fragment>
    <FormGroup label="Field" fieldId="field" isRequired
               helperText="The new field name that will be created on the gateway schema">
      <TextInput
        id="field" name="field"
        value={action.field} onChange={onChange.field}
        isRequired type="text"
      />
    </FormGroup>

    <FormGroup label="Upstream Server" isRequired fieldId="upstream" helperText="The upstream sever to send the request to">
      <BetterSelect value={action.upstream} setValue={onChange.upstream}>
        {upstreams.map((item, key) =>
          <SelectOption value={item.name} key={key}>{item.name}</SelectOption>
        )}
      </BetterSelect>
    </FormGroup>

    <FormGroup label="Query" isRequired fieldId="query" helperText="A partial upstream GraphQL query that requests will be routed to">
      <TextArea
        id="query" name="query"
        value={action.query} onChange={onChange.query}
        isRequired type="text"
      />
    </FormGroup>

    <VarList value={action.vars} onChange={onChange.vars}>test</VarList>

  </React.Fragment>
}

export default Link