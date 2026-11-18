// Empties internal/web/dist before a build, keeping the placeholder that makes `go:embed` always match.
import { readdirSync, rmSync } from 'node:fs';
import { fileURLToPath } from 'node:url';

const dir = fileURLToPath(new URL('../../internal/web/dist/', import.meta.url));
for (const f of readdirSync(dir)) if (f !== 'placeholder.txt') rmSync(dir + f, { recursive: true, force: true });
