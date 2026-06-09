import React, { ChangeEvent, useCallback } from 'react';

import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { InlineField, InlineFieldRow, Input, SecretInput } from '@grafana/ui';

import { DEFAULT_BASE_URL, MyDataSourceOptions, MySecureJsonData } from './types';

type Props = DataSourcePluginOptionsEditorProps<MyDataSourceOptions, MySecureJsonData>;

/**
 * ConfigEditor renders the data source configuration editor.
 *
 * It allows configuring:
 * - The base URL for the TAP service (defaults to the ESA Gaia archive)
 * - An optional API key for authenticated TAP services
 */
export function ConfigEditor({ options, onOptionsChange }: Props) {
  const { jsonData, secureJsonFields, secureJsonData } = options;

  const onBaseUrlChange = useCallback(
    (event: ChangeEvent<HTMLInputElement>) => {
      onOptionsChange({
        ...options,
        jsonData: {
          ...jsonData,
          baseUrl: event.target.value,
        },
      });
    },
    [options, jsonData, onOptionsChange]
  );

  const onApiKeyChange = useCallback(
    (event: ChangeEvent<HTMLInputElement>) => {
      onOptionsChange({
        ...options,
        secureJsonData: {
          ...secureJsonData,
          apiKey: event.target.value,
        },
      });
    },
    [options, secureJsonData, onOptionsChange]
  );

  const onResetApiKey = useCallback(() => {
    onOptionsChange({
      ...options,
      secureJsonFields: {
        ...secureJsonFields,
        apiKey: false,
      },
      secureJsonData: {
        ...secureJsonData,
        apiKey: '',
      },
    });
  }, [options, secureJsonFields, secureJsonData, onOptionsChange]);

  const isApiKeySet = Boolean(secureJsonFields?.apiKey);

  return (
    <div className="gf-form-group">
      <h3 className="page-heading">TAP Service Connection</h3>

      <InlineFieldRow>
        <InlineField
          label="TAP Service URL"
          labelWidth={20}
          tooltip={`Base URL for the IVOA TAP (Table Access Protocol) service. Defaults to the ESA Gaia archive at ${DEFAULT_BASE_URL}`}
          grow
        >
          <Input
            value={jsonData.baseUrl || ''}
            onChange={onBaseUrlChange}
            placeholder={DEFAULT_BASE_URL}
            width={60}
          />
        </InlineField>
      </InlineFieldRow>

      <h3 className="page-heading">Authentication (Optional)</h3>

      <InlineFieldRow>
        <InlineField
          label="API Key"
          labelWidth={20}
          tooltip="Optional API key for authenticated TAP services. Leave empty for public access to the Gaia archive."
        >
          <SecretInput
            isConfigured={isApiKeySet}
            value={secureJsonData?.apiKey || ''}
            placeholder="Optional API key"
            width={40}
            onReset={onResetApiKey}
            onChange={onApiKeyChange}
          />
        </InlineField>
      </InlineFieldRow>

      <div className="gf-form" style={{ marginTop: '8px' }}>
        <span className="gf-form-hint">
          <span className="gf-form-hint-icon">
            <i className="fa fa-info-circle" />
          </span>
          The public Gaia archive at{' '}
          <a href="https://gea.esac.esa.int/archive/" target="_blank" rel="noreferrer">
            https://gea.esac.esa.int/archive/
          </a>{' '}
          does not require authentication. You can use{' '}
          <a href="https://www.ivoa.net/documents/ADQL/" target="_blank" rel="noreferrer">
            ADQL
          </a>{' '}
          to query the Gaia DR3 catalog and other datasets.
        </span>
      </div>
    </div>
  );
}
