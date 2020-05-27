import * as React from 'react';

export function accessibleRouteChangeHandler() {
    return window.setTimeout(() => {
        const mainContainer = document.getElementById('primary-app-container');
        if (mainContainer) {
            mainContainer.focus();
        }
    }, 50);
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