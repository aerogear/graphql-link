import * as React from 'react';

import { accessibleRouteChangeHandler } from './utils';

class DynamicImport extends React.Component {
  state = {
    component: null,
  };
  routeFocusTimer = 0;
  
  constructor(props) {
    super(props);
    this.routeFocusTimer = 0;
  }
  
  componentWillUnmount() {
    window.clearTimeout(this.routeFocusTimer);
  }
  
  componentDidMount() {
    this.props
      .load()
      .then((component) => {
        if (component) {
          this.setState({
            component: component.default ? component.default : component,
          });
        }
      })
      .then(() => {
        if (this.props.focusContentAfterMount) {
          this.routeFocusTimer = accessibleRouteChangeHandler();
        }
      });
  }
  render() {
    return this.props.children(this.state.component);
  }
}

export { DynamicImport };