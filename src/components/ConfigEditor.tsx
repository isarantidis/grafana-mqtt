import React, { ChangeEvent, SyntheticEvent } from 'react';
import { InlineField, Input, SecretInput } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps, SelectableValue } from '@grafana/data';
import { MyDataSourceOptions, MySecureJsonData } from '../types';
import { QosSelect } from './QosSelect';

interface Props extends DataSourcePluginOptionsEditorProps<MyDataSourceOptions> { }

const LABEL_WIDTH = 20
export function ConfigEditor(props: Props) {
  const { onOptionsChange, options } = props;

  // Secure field (only sent to the backend)
  const onPasswordChange = (event: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      secureJsonData: {
        password: event.target.value,
      },
    });
  };

  const onResetPassword = () => {
    onOptionsChange({
      ...options,
      secureJsonFields: {
        ...options.secureJsonFields,
        password: false,
      },
      secureJsonData: {
        ...options.secureJsonData,
        password: '',
      },
    });
  };

  const { jsonData, secureJsonFields } = options;
  const secureJsonData = (options.secureJsonData || {}) as MySecureJsonData;

  return (
    <div className="gf-form-group">
      <>
        <h3 className="page-heading">Security</h3>
        <div className="gf-form-group">
          <InlineField label="Username" labelWidth={LABEL_WIDTH}>
            <Input
              onChange={onChangeHandler('username', options, onOptionsChange)}
              value={jsonData.username || ''}
              width={40}
              placeholder=''
            />
          </InlineField>
          <InlineField label="Password" labelWidth={LABEL_WIDTH}>
            <SecretInput
              isConfigured={(secureJsonFields && secureJsonFields.password) as boolean}
              value={secureJsonData.password || ''}
              placeholder="Password"
              width={40}
              onReset={onResetPassword}
              onChange={onPasswordChange}
            />
          </InlineField>
        </div>
        <h3 className="page-heading">Connection Settings</h3>
        <div className="gf-form-group">
          <InlineField label="Broker Url" labelWidth={LABEL_WIDTH}>
            <Input
              onChange={onChangeHandler('brokerUrl', options, onOptionsChange)}
              value={jsonData.brokerUrl || ''}
              width={40}
              placeholder='tcp://localhost:3181'
            />
          </InlineField>
          <InlineField label="Client Id" labelWidth={LABEL_WIDTH}>
            <Input
              onChange={onChangeHandler('clientId', options, onOptionsChange)}
              value={jsonData.clientId || ''}
              width={40}
              placeholder=''
            />
          </InlineField>
          
          <InlineField label="Quality of Service" labelWidth={LABEL_WIDTH}>
            <QosSelect value={jsonData.qos} width={40} updateValue={onChangeHandler('qos', options, onOptionsChange)}/>
          </InlineField>
        </div>
      </>
    </div>
  );
}

const getValueFromEventItem = (eventItem: SyntheticEvent<HTMLInputElement> | SelectableValue<any>) => {
  if (!eventItem) {
    return '';
  }

  if (eventItem.hasOwnProperty('currentTarget')) {
    return eventItem.currentTarget.value;
  }

  return (eventItem as SelectableValue<any>).value;
};

const onChangeHandler =
  (key: keyof MyDataSourceOptions, options: Props['options'], onOptionsChange: Props['onOptionsChange']) =>
    (eventItem: SyntheticEvent<HTMLInputElement> | SelectableValue<any>) => {
      onOptionsChange({
        ...options,
        jsonData: {
          ...options.jsonData,
          [key]: getValueFromEventItem(eventItem),
        },
      });
    };
