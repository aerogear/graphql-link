import './index.css';
import '@patternfly/react-core/dist/styles/base.css';

import React from 'react';
import { BrowserRouter as Router } from 'react-router-dom';

import AppLayout from './AppLayout';
import AppRoutes from './AppRoutes';

export default () => {
  return (
    <Router>
    <AppLayout>
      <AppRoutes />
    </AppLayout>
  </Router>
  );
};
