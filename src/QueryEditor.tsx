import { defaults } from 'lodash';
import React, { ChangeEvent, useCallback } from 'react';

import { QueryEditorProps } from '@grafana/data';
import { InlineField, InlineFieldRow, Input, TextArea } from '@grafana/ui';

import { GaiaADQLDataSource } from './datasource';
import { DEFAULT_QUERY, MyDataSourceOptions, MyQuery } from './types';

type Props = QueryEditorProps<GaiaADQLDataSource, MyQuery, MyDataSourceOptions>;

/**
 * QueryEditor renders the query editor for the Gaia ADQL data source.
 *
 * It provides:
 * - A textarea for entering ADQL queries
 * - A number input for limiting the number of returned rows
 */
export function QueryEditor({ query, onChange, onRunQuery }: Props) {
  const { adqlQuery, maxRows } = defaults(query, DEFAULT_QUERY);

  const onAdqlQueryChange = useCallback(
    (event: ChangeEvent<HTMLTextAreaElement>) => {
      onChange({ ...query, adqlQuery: event.target.value });
    },
    [query, onChange]
  );

  const onMaxRowsChange = useCallback(
    (event: ChangeEvent<HTMLInputElement>) => {
      const value = parseInt(event.target.value, 10);
      if (!isNaN(value) && value > 0) {
        onChange({ ...query, maxRows: value });
        onRunQuery();
      }
    },
    [query, onChange, onRunQuery]
  );

  const onAdqlQueryBlur = useCallback(() => {
    onRunQuery();
  }, [onRunQuery]);

  return (
    <div>
      <InlineFieldRow>
        <InlineField
          label="Max Rows"
          labelWidth={12}
          tooltip="Maximum number of rows to return from the Gaia TAP service. The TAP service may enforce its own limits."
        >
          <Input
            type="number"
            value={maxRows}
            onChange={onMaxRowsChange}
            width={12}
            min={1}
            max={5000000}
            step={100}
            placeholder="1000"
          />
        </InlineField>
      </InlineFieldRow>

      <InlineFieldRow>
        <InlineField
          label="ADQL Query"
          labelWidth={12}
          tooltip="Enter an ADQL (Astronomical Data Query Language) query to execute against the Gaia archive TAP service."
          grow
        >
          <TextArea
            value={adqlQuery}
            onChange={onAdqlQueryChange}
            onBlur={onAdqlQueryBlur}
            rows={10}
            placeholder={`SELECT TOP 10
  source_id, ra, dec, parallax, pmra, pmdec, phot_g_mean_mag
FROM gaiadr3.gaia_source
WHERE parallax > 10
ORDER BY phot_g_mean_mag ASC`}
            style={{ fontFamily: 'monospace', fontSize: '13px', width: '100%', minHeight: '200px' }}
          />
        </InlineField>
      </InlineFieldRow>
    </div>
  );
}
