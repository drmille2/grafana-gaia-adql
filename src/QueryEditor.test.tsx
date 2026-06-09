import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';

import { LoadingState, PanelData } from '@grafana/data';

import { GaiaADQLDataSource } from './datasource';
import { QueryEditor } from './QueryEditor';
import { MyQuery, DEFAULT_QUERY } from './types';

// Minimal mock for the data source
const mockDatasource = {} as GaiaADQLDataSource;

const defaultQuery: MyQuery = {
  refId: 'A',
  adqlQuery: DEFAULT_QUERY.adqlQuery!,
  maxRows: DEFAULT_QUERY.maxRows!,
};

const mockData: PanelData = {
  state: LoadingState.Done,
  series: [],
  timeRange: {
    from: {} as any,
    to: {} as any,
    raw: { from: 'now-6h', to: 'now' },
  },
};

describe('QueryEditor', () => {
  it('renders without crashing', () => {
    const onChange = jest.fn();
    const onRunQuery = jest.fn();

    render(
      <QueryEditor
        query={defaultQuery}
        datasource={mockDatasource}
        onChange={onChange}
        onRunQuery={onRunQuery}
        data={mockData}
        range={mockData.timeRange}
      />
    );

    expect(screen.getByPlaceholderText(/SELECT TOP 10/i)).toBeInTheDocument();
  });

  it('renders the Max Rows field', () => {
    const onChange = jest.fn();
    const onRunQuery = jest.fn();

    render(
      <QueryEditor
        query={defaultQuery}
        datasource={mockDatasource}
        onChange={onChange}
        onRunQuery={onRunQuery}
        data={mockData}
        range={mockData.timeRange}
      />
    );

    expect(screen.getByDisplayValue('1000')).toBeInTheDocument();
  });

  it('calls onChange when the ADQL query textarea is modified', () => {
    const onChange = jest.fn();
    const onRunQuery = jest.fn();

    render(
      <QueryEditor
        query={defaultQuery}
        datasource={mockDatasource}
        onChange={onChange}
        onRunQuery={onRunQuery}
        data={mockData}
        range={mockData.timeRange}
      />
    );

    const textarea = screen.getByRole('textbox', { name: /ADQL Query/i });
    fireEvent.change(textarea, { target: { value: 'SELECT TOP 5 source_id FROM gaiadr3.gaia_source' } });

    expect(onChange).toHaveBeenCalledWith(
      expect.objectContaining({
        adqlQuery: 'SELECT TOP 5 source_id FROM gaiadr3.gaia_source',
      })
    );
  });

  it('calls onRunQuery when the textarea loses focus', () => {
    const onChange = jest.fn();
    const onRunQuery = jest.fn();

    render(
      <QueryEditor
        query={defaultQuery}
        datasource={mockDatasource}
        onChange={onChange}
        onRunQuery={onRunQuery}
        data={mockData}
        range={mockData.timeRange}
      />
    );

    const textarea = screen.getByRole('textbox', { name: /ADQL Query/i });
    fireEvent.blur(textarea);

    expect(onRunQuery).toHaveBeenCalledTimes(1);
  });

  it('calls onChange with updated maxRows when the input changes', () => {
    const onChange = jest.fn();
    const onRunQuery = jest.fn();

    render(
      <QueryEditor
        query={defaultQuery}
        datasource={mockDatasource}
        onChange={onChange}
        onRunQuery={onRunQuery}
        data={mockData}
        range={mockData.timeRange}
      />
    );

    const maxRowsInput = screen.getByDisplayValue('1000');
    fireEvent.change(maxRowsInput, { target: { value: '500' } });

    expect(onChange).toHaveBeenCalledWith(
      expect.objectContaining({
        maxRows: 500,
      })
    );
  });

  it('renders with an empty/default query without errors', () => {
    const onChange = jest.fn();
    const onRunQuery = jest.fn();
    const emptyQuery: MyQuery = { refId: 'A', adqlQuery: '', maxRows: 100 };

    render(
      <QueryEditor
        query={emptyQuery}
        datasource={mockDatasource}
        onChange={onChange}
        onRunQuery={onRunQuery}
        data={mockData}
        range={mockData.timeRange}
      />
    );

    // Should still render the ADQL label
    expect(screen.getByText('ADQL Query')).toBeInTheDocument();
  });
});
