import React, { useState } from 'react';
import meta from './apbill.metadata.json';

export default function ApbillForm() {
  const [values, setValues] = useState({});

  const handleChange = (name, value) => {
    setValues(v => ({ ...v, [name]: value }));
  };

  const handleSubmit = e => {
    e.preventDefault();
    // TODO: wire up your save logic
    console.log('Submitted values:', values);
  };

  return (
    <form onSubmit={handleSubmit}>
      {meta.controls.map(ctrl => (
        <div key={ctrl.name} style={{ margin: '8px 0' }}>
          <label htmlFor={ctrl.name}>{ctrl.label}</label><br />
          {/* Render a control of type={ctrl.type} here */}
        </div>
      ))}
      <button type="submit">Submit</button>
    </form>
  );
}
