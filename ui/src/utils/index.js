import * as React from 'react';

export function accessibleRouteChangeHandler() {
  return window.setTimeout(() => {
    const mainContainer = document.getElementById('primary-app-container');
    if (mainContainer) {
      mainContainer.focus();
    }
  }, 50);
}

export function clone(v) {
  if (v === null || v === undefined) {
    return v
  }
  return JSON.parse(JSON.stringify(v))
}

export function chain(apply, then) {
  return function (...args) {
    try {
      return apply(...args)
    } finally {
      then(...args)
    }
  }
}

export function fieldSetters(source, setSource) {
  const rc = {}
  for (const field in source) {
    rc[field] = (x) => {
      const c = clone(source)
      c[field] = x
      setSource(c)
    }
  }
  return rc
}

export function setField(obj, setObj, key, value) {
  if (obj[key] === value) {
    const copy = clone(obj)
    copy[key] = value
    setObj(copy)
  }
}

// a custom hook for setting the page title
export function useDocumentTitle(title) {
  React.useEffect(() => {
    const originalTitle = document.title;
    document.title = title;

    return () => {
      document.title = originalTitle;
    };
  }, [title]);
}

export function toKeyedArray(map, by = "name", value = "value") {
  const v = []
  if (map) {
    for (const key in map) {
      const item = map[key]
      if (typeof (item) === "object") {
        item[by] = key
        v.push(item)
      } else {
        const c = {}
        c[by] = key
        c[value] = item
        v.push(c)
      }
    }
  }
  return v
}

export function fromKeyedArray(array, by = "name", value = null) {
  const v = {}
  if (array) {
    for (const item of array) {
      const key = item[by]
      if (value === null) {
        v[key] = item
      } else {
        v[key] = item[value]
      }
    }
  }
  return v
}