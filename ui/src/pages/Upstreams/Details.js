import {ActionGroup, Button, Flex, FlexItem, FlexModifiers, Form, FormGroup, TextInput} from '@patternfly/react-core';

import React from 'react';
import ConfirmDelete from "../../components/ConfirmDelete";
import {clone} from "../../utils";

const Details = ({onClose, id, value, setValue}) => {
  const upstream = value[id];
  const [name, setName] = React.useState(upstream.name);
  const [url, setUrl] = React.useState(upstream.url);
  const [prefix, setPrefix] = React.useState(upstream.prefix);
  const [suffix, setSuffix] = React.useState(upstream.suffix);

  const deleteDialog = ConfirmDelete((confirmed) => {
    if (confirmed) {
      value.splice(id, 1);
      setValue(value)
    }
    onClose()
  })

  const onSave = () => {
    const v = clone(value)
    v[id] = {
      name,
      url,
      prefix,
      suffix
    }
    setValue(v)
    onClose()
  }

  return (
    <Flex breakpointMods={[{modifier: FlexModifiers["space-items-lg"]}, {modifier: FlexModifiers["column"]}]}>
      <FlexItem>
        <Form>
          <FormGroup label="Name" isRequired fieldId="name" helperText="Upstream server name">
            <TextInput
              id="name" name="name"
              value={name} onChange={setName}
              isRequired type="text"
            />
          </FormGroup>
          <FormGroup label="URL" isRequired fieldId="name" helperText="URL to the GraphQL endpoint">
            <TextInput
              id="name" name="name"
              value={url} onChange={setUrl}
              isRequired type="text"
            />
          </FormGroup>
          <FormGroup label="Type Prefix" fieldId="name" helperText="Prefix to add to all this server's GraphQL types">
            <TextInput
              id="prefix" name="prefix"
              value={prefix} onChange={setPrefix}
              type="text"
            />
          </FormGroup>
          <FormGroup label="Type Suffix" fieldId="suffix" helperText="Suffix to add to all this server's GraphQL types">
            <TextInput
              id="suffix" name="suffix"
              value={suffix} onChange={setSuffix}
              type="text"
            />
          </FormGroup>
          <ActionGroup>
            <Button variant="primary" onClick={onSave}>Save</Button>
            <Button variant="secondary" onClick={onClose}>Cancel</Button>
            <Button variant="secondary" onClick={deleteDialog.open}>Delete</Button>
            {deleteDialog.render()}
          </ActionGroup>
        </Form>
      </FlexItem>
    </Flex>
  )
};
export default Details