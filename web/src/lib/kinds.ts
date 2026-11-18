/** What kind of device it is, as guessed by the server, with a small icon. */
const icons: Record<string, string> = {
  router: '📡',
  network: '🔌',
  computer: '💻',
  mobile: '📱',
  phone: '📱',
  tablet: '📱',
  tv: '📺',
  speaker: '🔊',
  printer: '🖨️',
  camera: '📷',
  nas: '🗄️',
  console: '🎮',
  iot: '💡',
  server: '🖥️',
};

export function kindIcon(kind: string): string {
  return icons[kind] ?? '❔';
}

export const knownKinds = Object.keys(icons);
