import React from "react";
import {
  ActionGroup,
  Breadcrumb,
  BreadcrumbItem,
  Button,
  Card,
  CardBody,
  Divider,
  Form,
  FormGroup,
  PageSection,
  PageSectionVariants,
  TextContent,
  TextInput,
} from "@patternfly/react-core";
import ActionList from "./ActionList";
import {Link, useHistory, useRouteMatch} from "react-router-dom";
import ConfirmDelete from "../../components/ConfirmDelete";
import {chain, clone, fieldSetters} from "../../utils";

const Type = ({config, onStoreConfig}) => {

  const history = useHistory()
  const match = useRouteMatch();
  const index = config.types.map(function (_) {
    return _.name;
  }).indexOf(match.params.typeName);

  const type = config.types[index];
  const storeType = (type)=>{
    const c = clone(config)
    c.types[index] = type
    onStoreConfig(c)
  }

  const store = fieldSetters(Object.assign({
    name: "",
    actions: [],
  }, type), storeType)


  const [name, setName] = React.useState(type.name);
  const reset = () => {
    setName(type.name)
  }

  const actions = type.actions;
  const setActions = store.actions


  const onSave = () => {
    store.name(name)
    history.push("/schema/types")
  }

  const deleteDialog = ConfirmDelete((confirmed) => {
    if (confirmed) {
      const c = clone(config)
      c.types.splice(index, 1)
      onStoreConfig(c)
      history.push("/schema/types")
    }
  })

  return <React.Fragment>
    <PageSection variant={PageSectionVariants.light}>
      <TextContent>
        <h1>{name}</h1>
      </TextContent>
      <Breadcrumb>
        <BreadcrumbItem>
          <Link to={`/schema/types`}>
            Types
          </Link>
        </BreadcrumbItem>
        <BreadcrumbItem isActive>
          <Link to={`${match.url}`}>
            {name}
          </Link>
        </BreadcrumbItem>
      </Breadcrumb>
    </PageSection>
    <Divider component="div"/>

    <PageSection>
      <Card>
        <CardBody>
          <Form>
            <FormGroup label="Type Name" isRequired fieldId="name"
                       helperText="Name of the type on the gateway (after any renames)">
              <TextInput
                id="name" name="name"
                value={name} onChange={setName}
                isRequired type="text"
              />
            </FormGroup>

            <ActionGroup>
              {
                name !== type.name && (<React.Fragment>
                  <Button variant="primary" onClick={onSave}>Save</Button>
                  <Button variant="secondary" onClick={reset}>Reset</Button>
                </React.Fragment>)
              }
              <Button variant="secondary" onClick={deleteDialog.open}>Delete</Button>{deleteDialog.render()}
            </ActionGroup>

          </Form>
        </CardBody>
      </Card>
    </PageSection>
    <PageSection>
      <ActionList config={config} actions={actions} setActions={setActions}/>
    </PageSection>
  </React.Fragment>
}

export default Type