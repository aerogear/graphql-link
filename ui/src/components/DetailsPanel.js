import React from 'react';
import ReactDOM from 'react-dom';
import './DetailsPanel.css'
import { TimesIcon } from '@patternfly/react-icons';

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
    <div className="pf-c-drawer pf-m-expanded">
      <div className="pf-c-drawer__main">
        <div className="pf-c-drawer__content">
          <div className="pf-c-drawer__body"></div>
        </div>
        <div className="pf-c-drawer__panel" aria-expanded="true">
          <div className="pf-c-drawer__body">
            <div className="pf-c-drawer__head">
              {children}
            </div>
          </div>
        </div>
      </div>
    </div>
</div>
  return ReactDOM.createPortal(content, elRef.current);
}

export const DetailsClose = ({onClick})=>{
  return <div className="pf-c-drawer__actions">
    <div className="pf-c-drawer__close" onClick={onClick}>
      <button className="pf-c-button pf-m-plain" type="button" aria-label="Close drawer panel">
        <TimesIcon/>
      </button>
    </div>
  </div>
}

export default DetailsPanel