/** Groups the risky services the server reports, so each group is explained once. */
const groups: Record<string, string[]> = {
  remoteAdmin: ['telnet', 'rdp', 'vnc', 'adb', 'docker-api', 'kubernetes', 'webmin'],
  database: ['redis', 'memcached', 'mongodb', 'elasticsearch', 'couchdb', 'mysql', 'postgres', 'mssql', 'oracle'],
  fileSharing: ['ftp', 'smb', 'rsync', 'nfs', 'netbios', 'msrpc', 'rpcbind'],
  legacy: ['snmp', 'socks', 'pptp'],
};

export function riskGroup(service: string): string {
  for (const [g, list] of Object.entries(groups)) if (list.includes(service)) return g;
  return 'legacy';
}
