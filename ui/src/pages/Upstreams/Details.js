import {
  ActionGroup,
  Button,
  Flex,
  FlexItem,
  FlexModifiers,
  Form,
  FormGroup,
  SelectOption,
  TextInput
} from '@patternfly/react-core';

import React from 'react';
import ConfirmDelete from "../../components/ConfirmDelete";
import {clone, fieldSetters} from "../../utils";
import BetterSelect from "../../components/BetterSelect";
import GraphQLUpstream from "./GraphQLUpstream";
import OpenAPIUpstream from "./OpenAPIUpstream";

const Details = ({onClose, id, value, setValue}) => {

  const initialState = clone(value[id] || {});
  initialState.type = initialState.type || "graphql"

  const [upstream, setUpstream] = React.useState(initialState);
  const onChange = fieldSetters(upstream, setUpstream)

  const deleteDialog = ConfirmDelete((confirmed) => {
    if (confirmed) {
      value.splice(id, 1);
      setValue(value)
    }
    onClose()
  })

  const onSave = () => {
    const v = clone(value)
    v[id] = upstream
    setValue(v)
    onClose()
  }

  return (
    <Flex breakpointMods={[{modifier: FlexModifiers["space-items-lg"]}, {modifier: FlexModifiers["column"]}]}>
      <FlexItem>
        <Form>
          <FormGroup label="Type" isRequired fieldId="type" helperText="The type of upstream server">
            <BetterSelect value={upstream.type} setValue={onChange.type}>
              <SelectOption value="graphql">GraphQL</SelectOption>
              <SelectOption value="openapi">OpenAPI</SelectOption>
            </BetterSelect>
          </FormGroup>

          <FormGroup label="Name" isRequired fieldId="name" helperText="Upstream server name">
            <TextInput
              id="name" name="name"
              value={upstream.name} onChange={onChange.name}
              isRequired type="text"
            />
          </FormGroup>

          {upstream.type === "graphql" && <GraphQLUpstream upstream={upstream} setUpstream={setUpstream}/>}
          {upstream.type === "openapi" && <OpenAPIUpstream upstream={upstream} setUpstream={setUpstream}/>}

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