import { render } from '@testing-library/react';
import React from 'react';

import App from '.';

test('render GraphQL Gateway Header', () => {
  const { getByText } = render(<App />);
  const linkElement = getByText(/GraphQL Gateway/i);
  expect(linkElement).toBeInTheDocument();
});
