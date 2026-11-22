// Fictitious data for the documentation screenshots (never a real network).
const ago = (min) => new Date(Date.now() - min * 60_000).toISOString();

export const devices = [
  { mac: 'a8:11:22:00:00:01', ip: '192.168.1.1', name: 'router.lan', vendor: 'Acme Networks Ltd', randomized_mac: false, online: true, first_seen: ago(60 * 24 * 40), last_seen: ago(1), label: '', tags: ['rede'], known: true, model: 'AX3000', kind: 'router', services: [], this_machine: false },
  { mac: 'b8:27:eb:00:00:0a', ip: '192.168.1.10', name: 'pi.local', vendor: 'Raspberry Pi Trading Ltd', randomized_mac: false, online: true, first_seen: ago(60 * 24 * 30), last_seen: ago(1), label: 'Servidor da sala', tags: ['servidor', 'iot'], known: true, model: '', kind: 'server', services: ['_ssh._tcp', '_http._tcp'], this_machine: true },
  { mac: 'da:44:55:00:00:14', ip: '192.168.1.20', name: 'Marias-iPhone.local', vendor: '', randomized_mac: true, online: true, first_seen: ago(60 * 24 * 12), last_seen: ago(1), label: 'iPhone da Maria', tags: ['celular'], known: true, model: 'iPhone14,5', kind: 'phone', services: ['_companion-link._tcp'] },
  { mac: 'a8:33:44:00:00:1f', ip: '192.168.1.31', name: '', vendor: 'Samsung Electronics Co.,Ltd', randomized_mac: false, online: true, first_seen: ago(60 * 3), last_seen: ago(1), label: '', tags: [], known: false, model: '', kind: '', services: [], this_machine: false },
  { mac: '3c:5a:b4:00:00:28', ip: '192.168.1.40', name: 'living-room-tv', vendor: 'Google, Inc.', randomized_mac: false, online: true, first_seen: ago(60 * 24 * 20), last_seen: ago(1), label: '', tags: [], known: false, model: 'Chromecast Ultra', kind: 'tv', services: ['_googlecast._tcp'] },
  { mac: 'ae:66:77:00:00:33', ip: '192.168.1.52', name: 'Joao-laptop.local', vendor: '', randomized_mac: true, online: false, first_seen: ago(60 * 24 * 9), last_seen: ago(95), label: '', tags: [], known: false, model: '', kind: 'computer', services: [], this_machine: false },
  { mac: 'a8:55:66:00:00:41', ip: '192.168.1.60', name: 'printer.lan', vendor: 'Brother Industries, Ltd.', randomized_mac: false, online: false, first_seen: ago(60 * 24 * 35), last_seen: ago(60 * 26), label: '', tags: [], known: false, model: 'HL-L2350DW', kind: 'printer', services: ['_ipp._tcp'] },
];

export const events = [
  { at: ago(40), kind: 'gateway_changed', detail: '192.168.1.1: a8:11:22:00:00:01 -> a8:99:99:00:00:09', mac: 'a8:11:22:00:00:01', name: 'router.lan', vendor: 'Acme Networks Ltd' },
  { at: ago(95), kind: 'offline', detail: '', mac: 'ae:66:77:00:00:33', name: 'Joao-laptop.local', vendor: '' },
  { at: ago(3 * 60), kind: 'joined', detail: '192.168.1.31', mac: 'a8:33:44:00:00:1f', name: '', vendor: 'Samsung Electronics Co.,Ltd' },
  { at: ago(60 * 5), kind: 'ip_changed', detail: '192.168.1.44 -> 192.168.1.52', mac: 'ae:66:77:00:00:33', name: 'Joao-laptop.local', vendor: '' },
  { at: ago(60 * 26), kind: 'offline', detail: '', mac: 'a8:55:66:00:00:41', name: 'printer.lan', vendor: 'Brother Industries, Ltd.' },
  { at: ago(60 * 30), kind: 'online', detail: '192.168.1.20', mac: 'da:44:55:00:00:14', name: 'Marias-iPhone.local', vendor: '' },
  { at: ago(60 * 38), kind: 'offline', detail: '', mac: 'da:44:55:00:00:14', name: 'Marias-iPhone.local', vendor: '' },
  { at: ago(60 * 24 * 9), kind: 'joined', detail: '192.168.1.44', mac: 'ae:66:77:00:00:33', name: 'Joao-laptop.local', vendor: '' },
];

export const ports = {
  at: new Date(Date.now() - 12 * 60_000).toISOString(),
  ports: [
    { port: 22, service: 'ssh' },
    { port: 80, service: 'http' },
    { port: 8080, service: 'http-proxy' },
  ],
  opened: [8080],
};
export const exposure = {
  at: new Date(Date.now() - 5 * 60_000).toISOString(),
  risks: [
    { port: 23, service: 'telnet', severity: 'high' },
    { port: 445, service: 'smb', severity: 'medium' },
  ],
  web: [
    { port: 80, tls: false, status: 200, server: 'lighttpd/1.4.71', title: 'Servidor da sala', auth: false },
    { port: 8443, tls: true, status: 401, server: '', title: 'Painel de administração', auth: true },
  ],
  tls: [{ port: 8443, version: 'TLS 1.3', subject: 'CN=pi.local', issuer: 'CN=pi.local', not_after: '2027-03-01T00:00:00Z', days_left: 102, expired: false, self_signed: true, legacy_tls: false }],
};
export const dhcp = {
  names: { enabled: true, active: true, learned: 3 },
  enabled: true,
  servers: [{ ip: '192.168.1.1', first_seen: new Date(Date.now() - 3 * 86400_000).toISOString(), last_seen: new Date(Date.now() - 4 * 60_000).toISOString(), router: '192.168.1.1', dns: ['192.168.1.1', '1.1.1.1'] }],
  last: { at: new Date(Date.now() - 4 * 60_000).toISOString(), offers: [], baseline: false, new: [] },
};
export const audit = [
  { at: new Date(Date.now() - 20 * 60_000).toISOString(), action: 'traceroute', target: '1.1.1.1', detail: '9 hops' },
  { at: new Date(Date.now() - 12 * 60_000).toISOString(), action: 'portscan', target: 'b8:27:eb:00:00:0a', detail: '192.168.1.10: 3 open' },
  { at: new Date(Date.now() - 50 * 60_000).toISOString(), action: 'ping', target: '192.168.1.1', detail: '4 sent, 0 lost' },
  { at: new Date(Date.now() - 26 * 3600_000).toISOString(), action: 'wake', target: 'a8:55:66:00:00:41', detail: 'magic packet to 192.168.1.255' },
];
export const settings = { retention_days: 90, scan_interval: '5m0s', port_scan_hours: 24, prune_days: 30, active_tools: true };
export const status = (extra = {}) => ({ version: 'v0.6.0', setup_required: false, authenticated: true, scan: { scanning: false, last_scan: ago(1), last_error: '', interval: '5m0s' }, ...extra });

export const alerts = {
  rules: { new_device: true, known_offline: false, arp_anomaly: true, rogue_dhcp: true, new_ports: true },
  channels: [
    { id: 'c1a2b3', type: 'ntfy', name: 'Celular (ntfy)', enabled: true, url: 'https://ntfy.sh/atalaia-exemplo', has_token: false, has_password: false },
    { id: 'd4e5f6', type: 'telegram', name: 'Telegram', enabled: true, chat_id: '123456789', has_token: true, has_password: false },
  ],
};

// Release notes as served by /api/releases (a short excerpt of releases.yaml).
export const releases = {
  current: 'v0.6.0',
  releases: [
    {
      version: '0.6.0', date: '2026-11-19',
      en: { title: 'Inventory housekeeping and security fixes', items: ['Forget a device, with its history, or let Atalaia forget untouched devices that have been offline for N days (opt-in).', 'Filter the dashboard by device type.', 'Export the inventory as CSV or JSON from the dashboard or with `atalaia export`; the CSV is safe to open in spreadsheets.', '`atalaia backup` makes a consistent, private copy of the database while the server runs.'] },
      'pt-BR': { title: 'Organização do inventário e correções de segurança', items: ['Esqueça um dispositivo, com o histórico, ou deixe o Atalaia esquecer os não mexidos que estão offline há N dias (opcional).', 'Filtre o painel por tipo de dispositivo.', 'Exporte o inventário em CSV ou JSON pelo painel ou com `atalaia export`; o CSV é seguro para abrir em planilhas.', '`atalaia backup` faz uma cópia consistente e privada do banco com o servidor rodando.'] },
    },
    {
      version: '0.5.0', date: '2026-11-18',
      en: { title: 'Better device identification', items: ['Devices are identified by mDNS service types and model, SSDP/UPnP and NetBIOS names.', 'A device-type guess (router, printer, TV, phone, camera…) with icons on the dashboard.'] },
      'pt-BR': { title: 'Identificação melhor dos dispositivos', items: ['Os dispositivos são identificados pelos tipos de serviço e modelo mDNS, por SSDP/UPnP e por nomes NetBIOS.', 'Palpite do tipo de dispositivo (roteador, impressora, TV, celular, câmera…) com ícones no painel.'] },
    },
  ],
};
