import {Checkbox, FormGroup} from '@patternfly/react-core';

import React from 'react';
import {fieldSetters} from "../../utils";
import NameValueList from "./NameValueList";
import StringList from "./StringList";

const Headers = ({headers, setHeaders}) => {

  headers = Object.assign({
    "disable-forwarding": false,
    set: [],
    remove: []
  }, headers)
  const onChange = fieldSetters(headers, setHeaders)

  return (
    <React.Fragment>
      <FormGroup>

        <Checkbox label="Forward client HTTP headers"
                  id="disable-forwarding" name="disable-forwarding"
                  aria-label="Should client HTTP headers get forwarded to the upstream server?"
                  isChecked={!headers["disable-forwarding"]} onChange={x => onChange["disable-forwarding"](!x)}/>

      </FormGroup>

      <NameValueList value={headers.set} onChange={onChange.set}
                     fieldId="headers"
                     label="Set Headers" helperText="Additional HTTP headers to set on the upstream request"
                     nameLabel="Header" valueLabel="Value"/>

      <StringList value={headers.remove} onChange={onChange.remove}
                  fieldId="remove"
                  label="Remove Headers" helperText="HTTP headers to remove from the upstream request"
      />

    </React.Fragment>
  )
};
export default Headers