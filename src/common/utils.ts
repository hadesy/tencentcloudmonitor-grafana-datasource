export function serviceGroupBy(
    services: Array<{ service: string; label: string; namespace: string; href: string; groupName?: string }>
) {
    const result = services.reduce((acc, cur) => {
        const {namespace, label, groupName = label} = cur;

        const existedGroup = acc.find((item) => item.label === groupName);
        if (!existedGroup) {
            acc.push({label: groupName, value: groupName, items: [{label, value: namespace}]});
            return acc;
        }

        existedGroup.items.push({label, value: namespace});
        return acc;
    }, [] as any[]);

    // 将只有一个子元素的项目进行特殊处理
    return result.map((item) =>
        item.items.length === 1 ? {label: item.items[0].label, value: item.items[0].value} : item
    );
}


export function hasVariable(data: string): boolean {
    return data.indexOf('$') === 0;
}

//由于接口返回的数据不统一，一些HACK
export function hackDimension(service: string, data: any[]): any {
    let dimensions: any[] = [];
    for (let item of data) {
        if (service === 'QCE/CVM' && item === 'vm_uuid') {
            item = 'InstanceId'
        }

        if (service === "QCE/CDB") {
            if (item === 'instanceid') {
                item = 'InstanceId'
            } else if (item === 'insttype') {
                item = 'InstanceType'
            }
        }

        dimensions.push({
            label: item,
            value: item,
        })
    }
    return dimensions
}
