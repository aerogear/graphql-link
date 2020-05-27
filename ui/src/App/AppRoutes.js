import { Alert, PageSection } from '@patternfly/react-core';
import * as React from 'react';
import { Route, Switch } from 'react-router-dom';
import { LastLocationProvider, useLastLocation } from 'react-router-last-location';

import { NotFound } from '../pages/NotFound';
import routes from '../routes';
import { accessibleRouteChangeHandler, useDocumentTitle } from '../utils';

// const getSupportModuleAsync = () => () => import(/* webpackChunkName: 'support' */ '@app/Support/Support');
// const Support = (routeProps) => {
//   const lastNavigation = useLastLocation();
//   return (
//     <DynamicImport load={getSupportModuleAsync()} focusContentAfterMount={lastNavigation !== null}>
//       {(Component) => {
//         let loadedComponent = null;
//         if (Component === null) {
//           loadedComponent = (
//             <PageSection aria-label="Loading Content Container">
//               <div className="pf-l-bullseye">
//                 <Alert title="Loading" className="pf-l-bullseye__item" />
//               </div>
//             </PageSection>
//           );
//         } else {
//           loadedComponent = <Component.Support {...routeProps} />;
//         }
//         return loadedComponent;
//       }}
//     </DynamicImport>
//   );
// };


let routeFocusTimer;

// a custom hook for sending focus to the primary content container
// after a view has loaded so that subsequent press of tab key
// sends focus directly to relevant content
const useA11yRouteChange = (isAsync) => {
  const lastNavigation = useLastLocation();
  React.useEffect(() => {
    if (!isAsync && lastNavigation !== null) {
      routeFocusTimer = accessibleRouteChangeHandler();
    }
    return () => {
      window.clearTimeout(routeFocusTimer);
    };
  }, [isAsync, lastNavigation]);
};

const RouteWithTitleUpdates = ({ component: Component, isAsync = false, title, ...rest }) => {
  useA11yRouteChange(isAsync);
  useDocumentTitle(title);

  function routeWithTitle(routeProps) {
    return <Component {...rest} {...routeProps} />;
  }

  return <Route render={routeWithTitle} />;
};

const PageNotFound = ({ title }) => {
  useDocumentTitle(title);
  return <Route component={NotFound} />;
};

export default () => (
  <LastLocationProvider>
    <Switch>
      {routes.map(({ path, exact, component, title, isAsync }, idx) => (
        <RouteWithTitleUpdates
          path={path}
          exact={exact}
          component={component}
          key={idx}
          title={title}
          isAsync={isAsync}
        />
      ))}
      <PageNotFound title="404 Page Not Found" />
    </Switch>
  </LastLocationProvider>
);