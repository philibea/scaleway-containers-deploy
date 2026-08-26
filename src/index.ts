import * as core from '@actions/core';
import { createClient, Client } from '@scaleway/sdk';
import { ENV, DEFAULTS } from './constants';
import { envOr, printOutputs } from './utils';
import { deploy, teardown } from './orchestrator';

function createClientWrapper(): Client {
  const accessKey = process.env[ENV.ACCESS_KEY];
  const secretKey = process.env[ENV.SECRET_KEY];
  
  if (!accessKey || !secretKey) {
    throw new Error('SCW_ACCESS_KEY and SCW_SECRET_KEY are required');
  }
  
  const client = createClient({
    accessKey,
    secretKey,
  });
  
  return client;
}

async function run(): Promise<void> {
  try {
    const pathRegistry = process.env[ENV.REGISTRY];
    const region = envOr(ENV.REGION, DEFAULTS.REGION);
    const type = envOr(ENV.TYPE, DEFAULTS.TYPE);
    
    if (!pathRegistry) {
      core.setFailed('SCW_REGISTRY is not set');
      return;
    }
    
    const client = createClientWrapper();
    
    if (type === 'deploy') {
      const result = await deploy(client, region, pathRegistry);
      
      printOutputs(
        result.container.domainName,
        result.domain?.hostname || `https://${result.container.domainName}`,
        result.container.id,
        result.container.namespaceId
      );
    } else if (type === 'teardown') {
      const deletedContainer = await teardown(client, region, pathRegistry);
      
      printOutputs(
        deletedContainer.domainName,
        `https://${deletedContainer.domainName}`,
        deletedContainer.id,
        deletedContainer.namespaceId
      );
    } else {
      core.setFailed(`Unknown type: ${type}. Valid types are: deploy, teardown`);
    }
  } catch (error) {
    core.setFailed(error instanceof Error ? error.message : 'An unknown error occurred');
  }
}

run();