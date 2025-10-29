import React from 'react';
import ReactDOM from 'react-dom/client';

const App = () => {
  return (
    <div style={{ padding: '20px' }}>
      <h1>Test success</h1>
      <p>Все работает</p>
      <ul>
        <li>test</li>
        <li>Test</li>
        <li>tEST</li>
      </ul>
    </div>
  );
};

const root = ReactDOM.createRoot(document.getElementById('root'));
root.render(<App />);