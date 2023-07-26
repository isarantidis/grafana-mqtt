import { SelectableValue } from '@grafana/data';
import { Select } from '@grafana/ui';
import React from 'react';
import { QoS } from 'types';

const ALO: SelectableValue<QoS> = {
  label: "At least once",
  value: QoS.AT_LEAST_ONCE
}
const AMO: SelectableValue<QoS> = {
  label: "At most once",
  value: QoS.AT_MOST_ONCE
}

const EO: SelectableValue<QoS> = {
  label: "Exactly once",
  value: QoS.EXACTLY_ONCE
}

const DEFAULT_VALUE = ALO;

interface QosSelectProps {
  updateValue: (qos: SelectableValue<QoS>) => void,
  value: QoS,
  width?: number
}

export const QosSelect = (props: QosSelectProps) => {
  return (
    <Select width={props.width || 20}
      options={[ALO, AMO, EO]}
      value={props.value || DEFAULT_VALUE}
      onChange={props.updateValue} />
  );
};
