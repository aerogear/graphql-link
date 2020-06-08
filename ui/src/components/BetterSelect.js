import {Select} from "@patternfly/react-core";
import React from "react";

const BetterSelect = ({value, setValue, expanded, setExpanded, ...rest}) => {

  let [_expanded, _setExpanded] = React.useState(expanded === undefined ? false : expanded);
  let [_value, _setValue]  = React.useState(value === undefined ? null : value);

  const onSelect = (event, value)=> {
    _setValue(value)
    setValue !== undefined && setValue(value)
    _setExpanded(false)
    setExpanded !== undefined && setExpanded(false)
  }

  const onToggle = (event, value)=> {
    _setExpanded(!_expanded)
    setExpanded !== undefined && setExpanded(value)
  }

  return (
    <Select
      selections={_value}
      placeholderText={"Please select ..."}
      onSelect={onSelect}
      isExpanded={_expanded}
      onToggle={onToggle}
      {...rest}/>
  )
};
export default BetterSelect