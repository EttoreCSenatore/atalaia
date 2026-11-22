// Types for fixtures.mjs (fictitious data shared by the screenshots and the public demo).
import type { Alerts, AuditEntry, Device, DeviceEvent, Dhcp, Exposure, PortsResult, Releases, Settings, Status } from '../src/lib/api';

export const devices: Device[];
export const events: DeviceEvent[];
export const ports: PortsResult;
export const exposure: Exposure;
export const dhcp: Dhcp;
export const audit: AuditEntry[];
export const settings: Settings;
export function status(extra?: Partial<Status>): Status;
export const alerts: Alerts;
export const releases: Releases;
