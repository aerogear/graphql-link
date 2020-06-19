import {
  ActionGroup,
  Button,
  Card,
  CardBody,
  DataList,
  DataListCell,
  DataListItem,
  DataListItemCells,
  DataListItemRow,
  EmptyState,
  Form,
  FormGroup,
  PageSection,
  PageSectionVariants,
  Split,
  SplitItem,
  TextContent,
  TextInput
} from '@patternfly/react-core';
import React from 'react';
import {Link} from "react-router-dom";
import {clone} from "../../utils";
import {Table, TableBody, TableHeader} from "@patternfly/react-table";
import DetailsPanel, {DetailsClose} from "../../components/DetailsPanel";

const SchemaDetails = ({schema, setSchema, onClose}) => {

  const [query, setQuery] = React.useState(schema.query || "Query")
  const [mutation, setMutation] = React.useState(schema.mutation || "Mutation")
  const [subscription, setSubscription] = React.useState(schema.subscription || "Subscription")

  const onSave = ()=>{
    const c = clone(schema)
    c.query = query
    c.mutation = mutation
    c.subscription = subscription
    setSchema(c)
  }

  return <Form>
    <FormGroup fieldId="query" label="Root Query Type" isRequired
               helperText="The type name of the root Query type">
      <TextInput
        id="query" name="query"
        value={query} onChange={setQuery}
        isRequired type="text"
      />
    </FormGroup>

    <FormGroup fieldId="mutation" label="Root Mutation Type" isRequired
               helperText="The type name of the root Mutation type">
      <TextInput
        id="mutation" name="mutation"
        value={mutation} onChange={setMutation}
        isRequired type="text"
      />
    </FormGroup>

    <FormGroup fieldId="subscription" label="Root Subscription Type" isRequired
               helperText="The type name of the root Subscription type">
      <TextInput
        id="subscription" name="subscription"
        value={subscription} onChange={setSubscription}
        isRequired type="text"
      />
    </FormGroup>
    <ActionGroup>
      <Button variant="primary" onClick={onSave}>Save</Button>
      <Button variant="secondary" onClick={onClose}>Cancel</Button>
    </ActionGroup>
  </Form>
}

const TypeList = ({config, onStoreConfig}) => {

  const [editRootTypes, setEditRootTypes] = React.useState(false)

  const value = config.types || []
  const schema = config.schema || {}

  const setSchema = (s) => {
    const c = clone(config)
    c.schema = s
    onStoreConfig(c)
    setEditRootTypes(false)
  }

  const onAdd = () => {
    const c = clone(config)
    c.types = c.types || []
    c.types.push({
      name: "Unnamed",
      actions: []
    })
    onStoreConfig(c)
  }

  return (
    <React.Fragment>
      <PageSection variant={PageSectionVariants.light}>
        <Split gutter="md">
          <SplitItem>
            <TextContent>
              <h1>Types</h1>
            </TextContent>
          </SplitItem>
          <SplitItem isFilled></SplitItem>
          <SplitItem><Button variant="primary" onClick={onAdd}>Add</Button></SplitItem>
        </Split>
      </PageSection>

      <PageSection>

        <Card>
          <CardBody>

            <Table aria-label="Root Types" cells={["Root Query", "Root Mutation", "Root Subscription"]} rows={[{
              cells:
                [schema.query, schema.mutation, schema.subscription]
            }]} actions={[{
              title: 'Edit',
              onClick: () => {
                setEditRootTypes(true)
              }
            },]}>
              <TableHeader/>
              <TableBody/>
            </Table>
            {
              editRootTypes &&
              <DetailsPanel id="right-panel">
                <DetailsClose onClick={() => setEditRootTypes(false)}></DetailsClose>
                <SchemaDetails schema={schema} setSchema={setSchema} onClose={() => setEditRootTypes(false)}/>
              </DetailsPanel>
            }
          </CardBody>
        </Card>
      </PageSection>

      <PageSection noPadding={false}>

        {value.length === 0 && <EmptyState>
          No types defined yet.
        </EmptyState>
        }
        {value.length !== 0 &&
        <DataList aria-label="data list">
          {value.map((item, key) =>
            <DataListItem id={"" + key} aria-labelledby={"" + key} key={key}>
              <DataListItemRow>
                <DataListItemCells
                  dataListCells={[
                    <DataListCell key="primary content">
                      <Link to={`./types/${item.name}`}>
                        {item.name}
                      </Link>
                    </DataListCell>,
                  ]}
                />
              </DataListItemRow>
            </DataListItem>
          )}
        </DataList>
        }

      </PageSection>

    </React.Fragment>
  )
};
export default TypeList