import {FormGroup, TextInput} from '@patternfly/react-core';

import React from 'react';
import {fieldSetters} from "../../utils";
import Headers from "./Headers";

const GraphQLUpstream = ({upstream, setUpstream}) => {

  upstream = Object.assign({
    suffix: "",
    prefix: "",
    url: "",
    headers: {},
  }, upstream)
  const onChange = fieldSetters(upstream, setUpstream)

  return (
    <React.Fragment>
      <FormGroup label="URL" isRequired fieldId="name" helperText="URL to the GraphQL endpoint">
        <TextInput
          id="url" name="url"
          value={upstream.url} onChange={onChange.url}
          isRequired type="text"
        />
      </FormGroup>
      <FormGroup label="Type Prefix" fieldId="name" helperText="Prefix to add to all this server's GraphQL types">
        <TextInput
          id="prefix" name="prefix"
          value={upstream.prefix} onChange={onChange.prefix}
          type="text"
        />
      </FormGroup>
      <FormGroup label="Type Suffix" fieldId="suffix" helperText="Suffix to add to all this server's GraphQL types">
        <TextInput
          id="suffix" name="suffix"
          value={upstream.suffix} onChange={onChange.suffix}
          type="text"
        />
      </FormGroup>

      <Headers headers={upstream.headers} setHeaders={onChange.headers}/>.
    </React.Fragment>
  )
};
export default GraphQLUpstream