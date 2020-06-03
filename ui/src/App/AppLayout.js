import {Nav, NavItem, NavList, Page, PageHeader, PageSidebar, SkipToContent} from '@patternfly/react-core';
import * as React from 'react';
import {NavLink} from 'react-router-dom';

import routes from '../routes';

export default (props) => {

  const logoProps = {
    href: '/',
    target: '_blank'
  };
  const [isNavOpen, setIsNavOpen] = React.useState(true);
  const [isMobileView, setIsMobileView] = React.useState(true);
  const [isNavOpenMobile, setIsNavOpenMobile] = React.useState(false);
  const onNavToggleMobile = () => {
    setIsNavOpenMobile(!isNavOpenMobile);
  };
  const onNavToggle = () => {
    setIsNavOpen(!isNavOpen);
  }
  const onPageResize = (props) => {
    setIsMobileView(props.mobileView);
  };
  const Header = (
    <PageHeader
      logo="GraphQL Gateway"
      logoProps={logoProps}
      showNavToggle
      isNavOpen={isNavOpen}
      onNavToggle={isMobileView ? onNavToggleMobile : onNavToggle}
    />
  );

  const Navigation = (
    <Nav id="nav-primary-simple" theme="dark">
      <NavList id="nav-list-simple">
        {routes.map((route, idx) => route.label && (
          <NavItem key={`${route.label}-${idx}`} id={`${route.label}-${idx}`}>
            <NavLink exact to={route.path} activeClassName="pf-m-current">{route.label}</NavLink>
          </NavItem>
        ))}
      </NavList>
    </Nav>
  );
  const Sidebar = (
    <PageSidebar
      theme="dark"
      nav={Navigation}
      isNavOpen={isMobileView ? isNavOpenMobile : isNavOpen}/>
  );
  const PageSkipToContent = (
    <SkipToContent href="#primary-app-container">
      Skip to Content
    </SkipToContent>
  );

  return (
    <Page
      mainContainerId="primary-app-container"
      header={Header}
      sidebar={Sidebar}
      onPageResize={onPageResize}
      skipToContent={PageSkipToContent}>
      {/*<div style={{background: "white"}}>*/}
        {props.children}
      {/*</div>*/}
    </Page>
  );
}