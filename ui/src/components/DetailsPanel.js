import React from 'react';
import ReactDOM from 'react-dom';
import './DetailsPanel.css'
import {TimesIcon} from '@patternfly/react-icons';
import {css} from '@patternfly/react-styles';
import styles from "@patternfly/react-styles/css/components/Drawer/drawer";
import {Button} from "@patternfly/react-core";

function getFreshDiv(id) {
  let el = document.getElementById(id)
  if (el) {
    el.parentElement.removeChild(el);
  }
  el = document.createElement('div');
  el.id = id
  return el
}


const DetailsPanel = ({id, children}) => {

  const elRef = React.useRef(getFreshDiv(id));
  React.useEffect(() => {
    const el = elRef.current
    document.body.appendChild(el);
    return () => {
      if (el.parentElement) {
        el.parentElement.removeChild(el);
      }
    }
  }, [id])

  const content = <div className="details-div">
    <div className={css(styles.drawer, styles.modifiers.expanded)}>
      <div className={css(styles.drawerMain)}>
        <div className={css(styles.drawerContent)}>
          <div className={css(styles.drawerBody)}></div>
        </div>
        <div className={css(styles.drawerPanel)} aria-expanded="true">
          <div className={css(styles.drawerBody)}>
            <div className={css(styles.drawerHead)}>
              {children}
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
  return ReactDOM.createPortal(content, elRef.current);
}

export const DetailsClose = ({onClick}) => {
  return <div className={css(styles.drawerActions)}>
    <div className={css(styles.drawerClose)} onClick={onClick}>
      <Button variant="plain"><TimesIcon/></Button>
    </div>
  </div>
}

export default DetailsPanel