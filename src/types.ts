import { DataSourceJsonData } from '@grafana/data';
import { DataQuery } from '@grafana/schema';

export interface MyQuery extends DataQuery {
  topic: string;
  includeSchema: boolean
  useInterval: boolean
}

export const DEFAULT_QUERY: Partial<MyQuery> = {
  topic: "topic",
  includeSchema: true,
  useInterval: false
};

/**
 * These are options configured for each DataSource instance
 */
export interface MyDataSourceOptions extends DataSourceJsonData {
  brokerUrl?: string
  clientId?: string
  username?: string
  qos: QoS
}

/**
 * Value that is used in the backend, but never sent over HTTP to the frontend
 */
export interface MySecureJsonData {
  password?: string;
}

export enum QoS {
  AT_LEAST_ONCE = 0,
  AT_MOST_ONCE = 1,
  EXACTLY_ONCE = 2
}
