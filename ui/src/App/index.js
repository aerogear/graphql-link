import './index.css';
import '@patternfly/react-core/dist/styles/base.css';

import React from 'react';
import {HashRouter as Router} from 'react-router-dom';

import AppLayout from './AppLayout';
import AppRoutes from './AppRoutes';

export default () => (
  <Router>
    <AppLayout>
      <AppRoutes/>
    </AppLayout>
  </Router>
)