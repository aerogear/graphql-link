import {Divider, PageSection, PageSectionVariants, Text, TextContent} from '@patternfly/react-core';
import React from 'react';
import UpstreamsList from "./UpstreamsList";
import UpstreamDetails from "./UpstreamDetails";
import HasDetails from "../../components/HasDetails";

export default () => {
  const [details, setDetails] = React.useState(null);

  return (
    <React.Fragment>
      <PageSection variant={PageSectionVariants.light}>
        <TextContent>
          <Text component="h1">Upstream Servers</Text>
        </TextContent>
      </PageSection>
      <Divider component="div"/>

      <PageSection noPadding={true}>
        <HasDetails details={details} components={{UpstreamDetails}}>
          <UpstreamsList details={details} setDetails={setDetails}/>
        </HasDetails>
      </PageSection>
    </React.Fragment>
  )
}