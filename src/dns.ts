import * as core from '@actions/core';
import { Container as ContainerSDK, Domain as DomainSDK, Client } from '@scaleway/sdk';
import { ENV, DNS } from './constants';

export async function deleteDnsRecord(
  client: Client,
  container: any,
  dnsZone: string
): Promise<any> {
  core.info('Update Zone DNS - Delete');
  
  const prefix = process.env[ENV.DNS_PREFIX] || '';
  const rootZone = process.env[ENV.ROOT_ZONE] || 'false';
  
  const api = new DomainSDK.v2beta1.API(client);
  
  const data = `${container.domainName}.`;
  
  let name = container.name;
  let type = DNS.CNAME;
  
  if (prefix) {
    name = prefix;
    core.info(`Update With Prefix Zone DNS - Delete: ${prefix}`);
  }
  
  if (rootZone === 'true') {
    name = '';
    type = 'ALIAS' as any;
    core.info('Update Root Zone DNS - Delete');
  }
  
  const changes = [
    {
      delete: {
        idFields: {
          name,
          data,
          type,
          ttl: DNS.TTL,
        },
      },
    },
  ];
  
  const response = await api.updateDNSZoneRecords({
    dnsZone,
    changes,
  } as any);
  
  return response;
}

export async function setDnsRecord(
  client: Client,
  container: any,
  dnsZone: string
): Promise<string> {
  const prefix = process.env[ENV.DNS_PREFIX] || '';
  const rootZone = process.env[ENV.ROOT_ZONE] || 'false';
  
  core.info('Update Zone DNS - Add');
  
  const api = new DomainSDK.v2beta1.API(client);
  
  let name = container.name;
  let type = DNS.CNAME;
  
  if (prefix) {
    name = prefix;
    core.info(`Update With Prefix Zone DNS - Add: ${prefix}`);
  }
  
  let hostname = `${name}.${dnsZone}`;
  
  if (rootZone === 'true') {
    name = '';
    type = 'ALIAS' as any;
    hostname = dnsZone;
    core.info('Update Root Zone DNS - Add');
  }
  
  const records = [
    {
      name,
      type,
      ttl: DNS.TTL,
      data: `${container.domainName}.`,
    },
  ];
  
  const data = `${container.domainName}.`;
  
  const changes = [
    {
      set: {
        idFields: {
          name,
          type,
          ttl: DNS.TTL,
          data,
        },
        records,
      },
    },
  ];
  
  await api.updateDNSZoneRecords({
    dnsZone,
    changes,
  } as any);
  
  core.info(`Hostname: ${hostname}`);
  
  return hostname;
}