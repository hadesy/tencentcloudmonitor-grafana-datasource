import React, {useEffect, useState} from 'react';
import {Button, Cascader, IconButton, InlineField, InlineFieldRow, Input, Select} from '@grafana/ui';
import {QueryEditorProps} from '@grafana/data';
import {DataSource} from '../datasource';
import {MyDataSourceOptions, MyQuery} from '../types';
import {SERVICES} from "../tc_monitor";
import {hackDimension, serviceGroupBy} from "../common/utils";
import _ from "lodash";

type Props = QueryEditorProps<DataSource, MyQuery, MyDataSourceOptions>;

export function QueryEditor({query, onChange, onRunQuery, datasource}: Props) {

    const [namespaces, setNamespaces] = useState<any[]>([])
    const [regions, setRegions] = useState<any[]>([])
    const [metrics, setMetrics] = useState<any[]>([])
    const [periods, setPeriods] = useState<any[]>([])
    const [dimensions, setDimensions] = useState<any[]>([])
    const [visible, setVisible] = useState<boolean>(false)

    const onCascaderSelect = (namespace: any) => {
        onChange({...query, service: namespace, period: 0, metric: '', dimensions: []});
    }

    useEffect(() => {
        setNamespaces(serviceGroupBy(SERVICES))
        setVisible(true)
    }, [])

    useEffect(() => {
        //更新地域
        if (_.isEmpty(query.service)) {
            return
        }
        datasource.getCvmRegions().then((value) => {
            setRegions(value)
        })
    }, [datasource, query.service])


    useEffect(() => {
        if (_.isEmpty(query.region)) {
            return
        }
        datasource.getBaseMetrics(query.service, query.region).then((value) => {
            setMetrics(value)
        })
    }, [query.service, datasource, query.region])

    useEffect(() => {
        if (_.isEmpty(query.metric)) {
            return
        }
        let selectedMetric = metrics.find(t => t.value === query.metric)
        if (_.isEmpty(selectedMetric)) {
            return
        }
        selectedMetric.period.sort((a: number, b: number) => a - b);
        let result: any[] = [];
        for (const item of selectedMetric.period) {
            result.push({
                label: item,
                value: item,
            })
        }
        setPeriods(result)
        setDimensions(hackDimension(selectedMetric.namespace, selectedMetric.dimensions[0].Dimensions))
    }, [metrics, query.metric])

    useEffect(() => {
        const {region, metric, period, dimensions, service} = query;
        console.log(query)

        if (!_.isEmpty(region) && !_.isEmpty(metric) && period !== 0
            && !_.isEmpty(dimensions) && !_.isEmpty(service)) {
            console.log(`onRunQuery`, query);
            onRunQuery();
        }

    }, [onRunQuery, query])


    return (
        <div>

            {visible && <InlineFieldRow>
                <InlineField label="Namespace">
                    <Cascader changeOnSelect={false} options={namespaces} onSelect={onCascaderSelect}
                              initialValue={query.service}/>
                </InlineField>
            </InlineFieldRow>}

            <InlineFieldRow>
                <InlineField label="Region">
                    <Select
                        options={regions}
                        value={query.region}
                        onChange={({value}) => {
                            onChange({...query, region: value as string});
                        }}/>
                </InlineField>
            </InlineFieldRow>


            <InlineFieldRow>
                <InlineField label="MetricName">
                    <Select
                        options={metrics}
                        value={query.metric}
                        onChange={({value}) => {
                            onChange({...query, metric: value as string});
                        }}/>
                </InlineField>
            </InlineFieldRow>


            <InlineFieldRow>
                <InlineField label="Period">
                    <Select
                        options={periods}
                        value={query.period}
                        onChange={({value}) => {
                            onChange({...query, period: value as number});
                        }}/>
                </InlineField>
            </InlineFieldRow>


            {!_.isEmpty(query.dimensions) && query.dimensions.map((field, index) => (
                <InlineFieldRow key={index}>
                    <InlineField label="Dimension">
                        <Select
                            options={dimensions}
                            value={query.dimensions[index].name}
                            onChange={({value}) => {
                                query.dimensions[index].name = value as string
                                onChange({...query});
                            }}/>
                    </InlineField>
                    <InlineField>
                        <Input value={query.dimensions[index].value} onChange={(event) => {
                            // @ts-ignore
                            query.dimensions[index].value = event.target.value
                            onChange({...query});
                        }}/>
                    </InlineField>
                    <InlineField style={{alignItems: "center"}}>
                        <IconButton name={"trash-alt"} size={"lg"} onClick={() => {
                            _.remove(query.dimensions, (value, index2) => index2 === index);
                            onChange({...query});
                        }}/>
                    </InlineField>
                </InlineFieldRow>
            ))}


            <Button onClick={() => {
                if (_.isNull(query.dimensions) || _.isUndefined(query.dimensions)) {
                    query.dimensions = []
                }
                query.dimensions.push({
                    name: '',
                    value: '',
                })
                onChange({...query});
            }}>Append Dimension</Button>


        </div>
    );
}
