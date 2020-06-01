import {Divider, PageSection, PageSectionVariants, Text, TextContent} from '@patternfly/react-core';
import React from 'react';
import UpstreamsList from "./UpstreamsList";
import UpstreamDetails from "./UpstreamDetails";
import ActionDetails from "./ActionDetails";
import ActionList from "./ActionList";
import HasDetails from "../../components/HasDetails";

export default () => {
  const [details, setDetails] = React.useState(null);

  return (
    <React.Fragment>
      <PageSection variant={PageSectionVariants.light}>
        <TextContent>
          <Text component="h1">Settings</Text>
          <Text component="p"></Text>
        </TextContent>
      </PageSection>
      <Divider component="div"/>

      <PageSection noPadding={true}>
        <HasDetails details={details} components={{UpstreamDetails, ActionDetails}}>
          <UpstreamsList details={details} setDetails={setDetails}/>
          <Divider component="div"/>
          <br/>
          <ActionList details={details} setDetails={setDetails}/>
        </HasDetails>
      </PageSection>

      {/*<UpstreamsList/>*/}
    </React.Fragment>
  )
}