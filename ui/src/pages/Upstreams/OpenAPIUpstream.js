import {Checkbox, FormGroup, TextInput} from '@patternfly/react-core';

import React from 'react';
import {fieldSetters} from "../../utils";

const OpenAPIUpstream = ({upstream, setUpstream}) => {

  upstream = Object.assign({
    suffix: "",
    prefix: "",
    spec: {},
    api: {},
  }, upstream)

  const onChange = fieldSetters(upstream, setUpstream)

  const spec = Object.assign({
    url: "",
    "insecure-client": false,
    "bearer-token": "",
  }, upstream.spec)

  const onChangeSpec = fieldSetters(spec, onChange.spec)

  const api = Object.assign({
    url: "",
    "insecure-client": false,
    "bearer-token": "",
    "api-key": ""
  }, upstream.api)
  const onChangeApi = fieldSetters(api, onChange.api)

  return (
    <React.Fragment>

      <FormGroup label="Spec URL" isRequired fieldId="spec.url"
                 helperText="URL to the OpenAPI specification document of the API">
        <TextInput
          id="spec.url" name="spec.url"
          value={spec.url} onChange={onChangeSpec.url}
          isRequired type="text"
        />
      </FormGroup>

      {
        spec.url.startsWith("https") &&
        <FormGroup fieldId="checkbox1">
          <Checkbox label="Disable https certificate verification" id="spec.insecure-client" name="spec.insecure-client"
                    aria-label="Disable https certificate verification"
                    isChecked={spec["insecure-client"]} onChange={onChangeSpec["insecure-client"]}/>
        </FormGroup>
      }

      <FormGroup label="API URL" isRequired fieldId="api.url"
                 helperText="URL to the API base (overrides the OpenAPI document setting)">
        <TextInput
          id="api.url" name="api.url"
          value={api.url} onChange={onChangeApi.url}
          isRequired type="text"
        />
      </FormGroup>

      {
        api.url.startsWith("https") &&
        <FormGroup fieldId="checkbox1">
          <Checkbox label="Disable https certificate verification" id="api.insecure-client" name="api.insecure-client"
                    aria-label="Disable https certificate verification"
                    isChecked={api["insecure-client"]} onChange={onChangeApi["insecure-client"]}/>
        </FormGroup>
      }

      <FormGroup label="API Key" isRequired fieldId="api.key"
                 helperText="URL to the API base (overrides the OpenAPI document setting)">
        <TextInput
          id="api.key" name="api.key"
          value={api["api-key"]} onChange={onChangeApi["api-key"]}
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

      <FormGroup label="API Bearer Token" isRequired fieldId="api.bearer-token"
                 helperText="Bearer token to place in the Authentication header of API requests">
        <TextInput
          id="api.bearer-token" name="api.bearer-token"
          value={api["bearer-token"]} onChange={onChangeApi["bearer-token"]}
          isRequired type="text"
        />
      </FormGroup>

      <FormGroup label="Spec Bearer Token" isRequired fieldId="spec.bearer-token"
                 helperText="Bearer token to place in the Authentication header when getting the OpenAPI specification document">
        <TextInput
          id="spec.bearer-token" name="spec.bearer-token"
          value={spec["bearer-token"]} onChange={onChangeSpec["bearer-token"]}
          isRequired type="text"
        />
      </FormGroup>


    </React.Fragment>
  )
};
export default OpenAPIUpstream