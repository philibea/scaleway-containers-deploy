import * as core from '@actions/core';
import { Client, Container as ContainerSDK } from '@scaleway/sdk';
import { ENV, DEFAULTS } from './constants';
import { envToInt, envOr, parseKeyValue, parseSecrets } from './utils';

export interface ContainerEnv {
  port: number;
  memoryLimit: number;
  minScale: number;
  maxScale: number;
  maxConcurrency: number;
  cpuLimit: number;
  sandbox: string;
}

export function getSandboxVersion(): string {
  const sandbox = envOr(ENV.SANDBOX, DEFAULTS.SANDBOX);
  
  if (sandbox === 'v1') {
    return 'v1';
  }
  
  if (sandbox === 'v2') {
    return 'v2';
  }
  
  return 'unknown';
}

export function getContainerEnvVariables(): ContainerEnv {
  return {
    port: envToInt(ENV.CONTAINER_PORT, DEFAULTS.PORT),
    memoryLimit: envToInt(ENV.MEMORY_LIMIT, DEFAULTS.MEMORY_LIMIT),
    minScale: envToInt(ENV.MIN_SCALE, DEFAULTS.MIN_SCALE),
    maxScale: envToInt(ENV.MAX_SCALE, DEFAULTS.MAX_SCALE),
    maxConcurrency: envToInt(ENV.MAX_CONCURRENCY, DEFAULTS.MAX_CONCURRENCY),
    cpuLimit: envToInt(ENV.CPU_LIMIT, DEFAULTS.CPU_LIMIT),
    sandbox: getSandboxVersion(),
  };
}

export async function waitForNamespaceReady(
  client: Client,
  namespace: any
): Promise<any> {
  core.info('Waiting for namespace to be ready...');
  
  const api = new ContainerSDK.v1beta1.API(client);
  
  const readyNamespace = await api.waitForNamespace({
    region: namespace.region,
    namespaceId: namespace.id,
  });
  
  return readyNamespace;
}

export async function waitForContainerReady(
  client: Client,
  container: any
): Promise<any> {
  core.info('Waiting for container to be ready...');
  
  const api = new ContainerSDK.v1beta1.API(client);
  
  const readyContainer = await api.waitForContainer({
    region: container.region,
    containerId: container.id,
  });
  
  return readyContainer;
}

export async function getContainer(
  client: Client,
  region: string,
  containerName: string
): Promise<any | null> {
  const namespaceId = process.env[ENV.CONTAINER_NAMESPACE_ID];
  
  if (!namespaceId) {
    throw new Error('Namespace ID not found');
  }
  
  const api = new ContainerSDK.v1beta1.API(client);
  
  const response = await api.listContainers({
    region,
    namespaceId,
    name: containerName,
  });
  
  if (response.containers.length === 0) {
    return null;
  }
  
  return response.containers[0];
}

export async function deleteContainer(
  client: Client,
  region: string,
  container: any
): Promise<any> {
  const api = new ContainerSDK.v1beta1.API(client);
  
  const deletedContainer = await api.deleteContainer({
    region,
    containerId: container.id,
  });
  
  return deletedContainer;
}

export async function getContainersNamespace(
  client: Client,
  region: string
): Promise<any> {
  const namespaceId = process.env[ENV.CONTAINER_NAMESPACE_ID];
  
  if (!namespaceId) {
    throw new Error('Containers namespace ID not found');
  }
  
  const api = new ContainerSDK.v1beta1.API(client);
  
  const namespace = await api.getNamespace({
    region,
    namespaceId,
  });
  
  return namespace;
}

export async function isContainerAlreadyCreated(
  client: Client,
  namespace: any,
  containerName: string
): Promise<any | null> {
  const api = new ContainerSDK.v1beta1.API(client);
  
  const response = await api.listContainers({
    region: namespace.region,
    namespaceId: namespace.id,
    name: containerName,
  });
  
  if (response.containers.length === 0) {
    return null;
  }
  
  return response.containers[0];
}

export async function updateDeployedContainer(
  client: Client,
  container: any,
  pathRegistry: string
): Promise<any> {
  const api = new ContainerSDK.v1beta1.API(client);
  
  const containerEnv = getContainerEnvVariables();
  const secrets = parseSecrets();
  const environmentVariables = parseKeyValue(process.env[ENV.ENVIRONMENT_VARIABLES] || '');
  
  const updatedContainer = await api.updateContainer({
    region: container.region,
    containerId: container.id,
    registryImage: pathRegistry,
    redeploy: true,
    environmentVariables,
    secretEnvironmentVariables: secrets,
    memoryLimit: containerEnv.memoryLimit,
    minScale: containerEnv.minScale,
    maxScale: containerEnv.maxScale,
    cpuLimit: containerEnv.cpuLimit,
    port: containerEnv.port,
    maxConcurrency: containerEnv.maxConcurrency,
    sandbox: containerEnv.sandbox,
  } as any);
  
  return updatedContainer;
}

export async function createContainerAndDeploy(
  client: Client,
  namespace: any,
  pathRegistry: string,
  containerName: string
): Promise<any> {
  const api = new ContainerSDK.v1beta1.API(client);
  
  const containerEnv = getContainerEnvVariables();
  const secrets = parseSecrets();
  const environmentVariables = parseKeyValue(process.env[ENV.ENVIRONMENT_VARIABLES] || '');
  
  const createdContainer = await api.createContainer({
    description: DEFAULTS.DESCRIPTION,
    name: containerName,
    namespaceId: namespace.id,
    region: namespace.region,
    registryImage: pathRegistry,
    timeout: `${DEFAULTS.TIMEOUT_SECONDS}s`,
    environmentVariables,
    secretEnvironmentVariables: secrets,
    memoryLimit: containerEnv.memoryLimit,
    minScale: containerEnv.minScale,
    maxScale: containerEnv.maxScale,
    cpuLimit: containerEnv.cpuLimit,
    port: containerEnv.port,
    maxConcurrency: containerEnv.maxConcurrency,
    sandbox: containerEnv.sandbox,
  } as any);
  
  const deployedContainer = await api.deployContainer({
    region: namespace.region,
    containerId: createdContainer.id,
  });
  
  return deployedContainer;
}

export async function deployContainer(
  client: Client,
  namespace: any,
  containerName: string,
  pathRegistry: string
): Promise<any> {
  core.info(`Container Name: ${containerName}`);
  
  const existingContainer = await isContainerAlreadyCreated(client, namespace, containerName);
  
  if (existingContainer) {
    core.info('Container already exists and will be updated');
    
    const updatedContainer = await updateDeployedContainer(client, existingContainer, pathRegistry);
    const readyContainer = await waitForContainerReady(client, updatedContainer);
    
    return readyContainer;
  } else {
    const newContainer = await createContainerAndDeploy(client, namespace, pathRegistry, containerName);
    const readyContainer = await waitForContainerReady(client, newContainer);
    
    return readyContainer;
  }
}

export async function setCustomDomainContainer(
  client: Client,
  container: any,
  hostname: string
): Promise<any> {
  if (!hostname) {
    throw new Error('Hostname is required');
  }
  
  if (hostname.length > 63) {
    throw new Error('Hostname cannot be longer than 63 characters');
  }
  
  const api = new ContainerSDK.v1beta1.API(client);
  
  const listResponse = await api.listDomains({
    region: container.region,
    containerId: container.id,
  });
  
  for (const domain of listResponse.domains) {
    if (domain.hostname === hostname) {
      return domain;
    }
  }
  
  const createdDomain = await api.createDomain({
    region: container.region,
    containerId: container.id,
    hostname,
  });
  
  return createdDomain;
}