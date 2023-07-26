import React, { ChangeEvent } from 'react';
import { Checkbox, InlineField, Input } from '@grafana/ui';
import { QueryEditorProps } from '@grafana/data';
import { DataSource } from '../datasource';
import { MyDataSourceOptions, MyQuery } from '../types';

type Props = QueryEditorProps<DataSource, MyQuery, MyDataSourceOptions>;

export function QueryEditor({ query, onChange, onRunQuery }: Props) {
  const onTopicChange = (event: ChangeEvent<HTMLInputElement>) => {
    onChange({ ...query, topic: event.target.value });
  };

  const onIncludeSchema = (event: ChangeEvent<HTMLInputElement>) => {
    onChange({ ...query, includeSchema: event.target.checked });
  };

  const onUseInterval = (event: ChangeEvent<HTMLInputElement>) => {
    onChange({ ...query, useInterval: event.target.checked });
  };

  return (
    <div className="gf-form">
      <InlineField label="Topic">
        <Input onChange={onTopicChange} value={query.topic} width={30} type="text" onBlur={onRunQuery} />
      </InlineField>
      <Checkbox label="Use interval" checked={query.useInterval} onBlur={onRunQuery} onChange={onUseInterval}/>
      <Checkbox label="Include Schema" checked={query.includeSchema} onBlur={onRunQuery} onChange={onIncludeSchema}/>
    </div>
  );
}
