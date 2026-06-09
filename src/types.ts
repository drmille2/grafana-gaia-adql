import { DataQuery, DataSourceJsonData } from '@grafana/data';

/**
 * Options configured for each DataSource instance.
 * These are stored in the plugin's jsonData field.
 */
export interface MyDataSourceOptions extends DataSourceJsonData {
  /** Base URL for the TAP service endpoint */
  baseUrl?: string;
}

/**
 * Value that is used in the backend, but never sent over HTTP to the frontend.
 * These fields are stored in the plugin's secureJsonData field.
 */
export interface MySecureJsonData {
  /** Optional API key for authenticated TAP services */
  apiKey?: string;
}

/**
 * Represents a single ADQL query panel configuration.
 */
export interface MyQuery extends DataQuery {
  /** The ADQL query string to execute against the Gaia TAP service */
  adqlQuery: string;

  /** Maximum number of rows to return from the TAP service */
  maxRows: number;
}

/**
 * Default values for a new query.
 */
export const DEFAULT_QUERY: Partial<MyQuery> = {
  adqlQuery: `SELECT TOP 10
  source_id, ra, dec, parallax, pmra, pmdec, phot_g_mean_mag
FROM gaiadr3.gaia_source
WHERE parallax > 10
ORDER BY phot_g_mean_mag ASC`,
  maxRows: 1000,
};

/**
 * Default TAP service URL for the Gaia archive.
 */
export const DEFAULT_BASE_URL = 'https://gea.esac.esa.int/tap-server/tap';
