// Fictitious data for the documentation screenshots (never a real network).
const ago = (min) => new Date(Date.now() - min * 60_000).toISOString();

export const devices = [
  { mac: 'a8:11:22:00:00:01', ip: '192.168.1.1', name: 'router.lan', vendor: 'Acme Networks Ltd', randomized_mac: false, online: true, first_seen: ago(60 * 24 * 40), last_seen: ago(1), label: '', tags: ['rede'], known: true },
  { mac: 'b8:27:eb:00:00:0a', ip: '192.168.1.10', name: 'pi.local', vendor: 'Raspberry Pi Trading Ltd', randomized_mac: false, online: true, first_seen: ago(60 * 24 * 30), last_seen: ago(1), label: 'Servidor da sala', tags: ['servidor', 'iot'], known: true },
  { mac: 'da:44:55:00:00:14', ip: '192.168.1.20', name: 'Marias-iPhone.local', vendor: '', randomized_mac: true, online: true, first_seen: ago(60 * 24 * 12), last_seen: ago(1), label: 'iPhone da Maria', tags: ['celular'], known: true },
  { mac: 'a8:33:44:00:00:1f', ip: '192.168.1.31', name: '', vendor: 'Samsung Electronics Co.,Ltd', randomized_mac: false, online: true, first_seen: ago(60 * 3), last_seen: ago(1), label: '', tags: [], known: false },
  { mac: '3c:5a:b4:00:00:28', ip: '192.168.1.40', name: 'living-room-tv', vendor: 'Google, Inc.', randomized_mac: false, online: true, first_seen: ago(60 * 24 * 20), last_seen: ago(1), label: '', tags: [], known: false },
  { mac: 'ae:66:77:00:00:33', ip: '192.168.1.52', name: 'Joao-laptop.local', vendor: '', randomized_mac: true, online: false, first_seen: ago(60 * 24 * 9), last_seen: ago(95), label: '', tags: [], known: false },
  { mac: 'a8:55:66:00:00:41', ip: '192.168.1.60', name: 'printer.lan', vendor: 'Brother Industries, Ltd.', randomized_mac: false, online: false, first_seen: ago(60 * 24 * 35), last_seen: ago(60 * 26), label: '', tags: [], known: false },
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
export const audit = [
  { at: new Date(Date.now() - 12 * 60_000).toISOString(), action: 'portscan', target: 'b8:27:eb:00:00:0a', detail: '192.168.1.10: 3 open' },
  { at: new Date(Date.now() - 50 * 60_000).toISOString(), action: 'ping', target: '192.168.1.1', detail: '4 sent, 0 lost' },
  { at: new Date(Date.now() - 26 * 3600_000).toISOString(), action: 'wake', target: 'a8:55:66:00:00:41', detail: 'magic packet to 192.168.1.255' },
];
export const settings = { retention_days: 90, scan_interval: '5m0s' };
export const status = (extra = {}) => ({ version: '0.2.0', setup_required: false, authenticated: true, scan: { scanning: false, last_scan: ago(1), last_error: '', interval: '5m0s' }, ...extra });

export const alerts = {
  rules: { new_device: true, known_offline: false, arp_anomaly: true },
  channels: [
    { id: 'c1a2b3', type: 'ntfy', name: 'Celular (ntfy)', enabled: true, url: 'https://ntfy.sh/atalaia-exemplo', has_token: false, has_password: false },
    { id: 'd4e5f6', type: 'telegram', name: 'Telegram', enabled: true, chat_id: '123456789', has_token: true, has_password: false },
  ],
};
