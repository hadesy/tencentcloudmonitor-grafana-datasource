import {DataSourceInstanceSettings, DataQueryRequest, DataQueryResponse} from '@grafana/data';
import {DataSourceWithBackend, getTemplateSrv} from '@grafana/runtime';
import {Observable} from 'rxjs';

import {MyQuery, MyDataSourceOptions} from './types';
import {hasVariable} from "./common/utils";

export class DataSource extends DataSourceWithBackend<MyQuery, MyDataSourceOptions> {
    constructor(instanceSettings: DataSourceInstanceSettings<MyDataSourceOptions>) {
        super(instanceSettings);
    }


    query(request: DataQueryRequest<MyQuery>): Observable<DataQueryResponse> {
        const templateSrv = getTemplateSrv();
        for (const item of request.targets) {
            for (const dimension of item.dimensions) {
                if (hasVariable(dimension.value)) {
                    dimension.value = templateSrv.replace(dimension.value, undefined, (result: any) => {
                        return result;
                    });
                }
            }
        }
        return super.query(request);
    }

    async getCvmRegions(): Promise<any[]> {
        let cvmRegions: any[] = (await this.getResource('cvm-regions')).cvmRegions
        let result: any[] = [];
        for (const item of cvmRegions) {
            result.push({
                label: item.RegionName,
                value: item.Region,
            })
        }
        return result
    }


    async getBaseMetrics(namespace: string, region: string): Promise<any[]> {
        let metrics: any[] = (await this.getResource('/monitor/describeBaseMetrics', {namespace, region})).metrics
        let result: any[] = [];
        for (const item of metrics) {
            result.push({
                label: item.MetricName + "(" + item.MetricCName + ")",
                value: item.MetricName,
                period: item.Period,
                dimensions: item.Dimensions,
                namespace: item.Namespace
            })
        }
        return result
    }

}
