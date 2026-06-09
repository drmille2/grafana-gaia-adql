import {
  DataQueryRequest,
  DataSourceInstanceSettings,
  ScopedVars,
  TestDataSourceResponse,
} from '@grafana/data';
import { DataSourceWithBackend, getTemplateSrv } from '@grafana/runtime';

import { MyDataSourceOptions, MyQuery } from './types';

/**
 * GaiaADQLDataSource is the frontend DataSource class that communicates
 * with the Go backend plugin to execute ADQL queries against the Gaia
 * astronomical catalog's TAP service.
 */
export class GaiaADQLDataSource extends DataSourceWithBackend<MyQuery, MyDataSourceOptions> {
  /** Base URL for the TAP service, stored in instance settings */
  readonly baseUrl: string;

  constructor(instanceSettings: DataSourceInstanceSettings<MyDataSourceOptions>) {
    super(instanceSettings);
    this.baseUrl = instanceSettings.jsonData.baseUrl || 'https://gea.esac.esa.int/tap-server/tap';
  }

  /**
   * Override query to apply template variable interpolation to ADQL queries
   * before sending them to the backend.
   */
  query(request: DataQueryRequest<MyQuery>): ReturnType<DataSourceWithBackend['query']> {
    const processedRequest = {
      ...request,
      targets: request.targets
        .filter((t) => !t.hide)
        .map((target) => this.applyTemplateVariables(target, request.scopedVars)),
    };

    return super.query(processedRequest);
  }

  /**
   * Applies Grafana template variable substitution to the ADQL query string.
   */
  applyTemplateVariables(query: MyQuery, scopedVars: ScopedVars): MyQuery {
    return {
      ...query,
      adqlQuery: getTemplateSrv().replace(query.adqlQuery, scopedVars),
    };
  }

  /**
   * Checks the health of the data source by calling the backend CheckHealth handler,
   * which in turn performs a simple TAP service availability check.
   */
  async testDatasource(): Promise<TestDataSourceResponse> {
    return super.testDatasource();
  }
}
