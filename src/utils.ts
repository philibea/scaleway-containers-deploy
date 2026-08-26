import * as core from '@actions/core';
import { CONTAINER_NAME_MAX_LENGTH } from './constants';

export function envOr(name: string, defaultValue: string): string {
  const value = process.env[name];
  return value !== undefined ? value : defaultValue;
}

export function envToInt(name: string, defaultValue: number): number {
  const value = envOr(name, defaultValue.toString());
  const parsed = parseInt(value, 10);
  return isNaN(parsed) ? defaultValue : parsed;
}

export function setOutput(name: string, value: string): void {
  core.setOutput(name, value);
}

export function printOutputs(containerUrl: string, url: string, containerId: string, namespaceId: string): void {
  setOutput('url', url);
  setOutput('container_url', containerUrl);
  setOutput('scw_container_id', containerId);
  setOutput('scw_namespace_id', namespaceId);
}

export function getContainerName(pathRegistry: string): string {
  const splitPath = pathRegistry.split(':');
  let name = splitPath[1] || '';
  
  name = name.replace(/-/g, '');
  name = name.replace(/_/g, '');
  
  if (name.length > CONTAINER_NAME_MAX_LENGTH) {
    name = name.substring(0, CONTAINER_NAME_MAX_LENGTH);
  }
  
  return name;
}

export function parseKeyValue(key: string): Record<string, string> {
  const keyValue: Record<string, string> = {};
  const envValue = process.env[key] || '';
  
  if (!envValue) {
    return keyValue;
  }
  
  const pairs = envValue.split(',');
  
  for (const pair of pairs) {
    const splitEnv = pair.split('=');
    
    if (splitEnv.length === 2) {
      keyValue[splitEnv[0]] = splitEnv[1];
    }
  }
  
  return keyValue;
}

export function parseSecrets(): Array<{ key: string; value: string }> {
  const secretsMap = parseKeyValue(process.env['INPUT_SCW_SECRETS'] || '');
  const secrets: Array<{ key: string; value: string }> = [];
  
  for (const [key, value] of Object.entries(secretsMap)) {
    secrets.push({ key, value });
  }
  
  return secrets;
}