import { DataSourcePlugin } from '@grafana/data';

import { ConfigEditor } from './ConfigEditor';
import { GaiaADQLDataSource } from './datasource';
import { QueryEditor } from './QueryEditor';
import { MyDataSourceOptions, MyQuery } from './types';

/**
 * Registers the Gaia ADQL data source plugin with Grafana.
 *
 * The plugin consists of:
 * - GaiaADQLDataSource: the data source class that handles communication
 *   with the Go backend plugin
 * - QueryEditor: the panel query editor component
 * - ConfigEditor: the data source configuration editor component
 */
export const plugin = new DataSourcePlugin<GaiaADQLDataSource, MyQuery, MyDataSourceOptions>(GaiaADQLDataSource)
  .setConfigEditor(ConfigEditor)
  .setQueryEditor(QueryEditor);
