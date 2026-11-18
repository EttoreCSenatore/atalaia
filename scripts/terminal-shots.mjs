// Renderiza arquivos .txt (comando na 1ª linha + saída da CLI) como imagens de terminal para a documentação.
// Uso: PLAYWRIGHT=/caminho/playwright/index.mjs node scripts/terminal-shots.mjs <pasta-com-txt> <pasta-de-saida>
// (cada nome.txt vira cli-nome.png). Os exemplos da documentação usam dados fictícios, nunca a rede real de alguém.
import {readdirSync, readFileSync} from 'node:fs';

const {chromium} = await import(process.env.PLAYWRIGHT ?? 'playwright');
const [, , entrada, saida] = process.argv;
const esc = (s) => s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
const browser = await chromium.launch();
for (const arq of readdirSync(entrada).filter((f) => f.endsWith('.txt'))) {
  const [cmd, ...resto] = readFileSync(`${entrada}/${arq}`, 'utf8').trimEnd().split('\n');
  const html = `<body style="margin:0;background:#0b1020"><div style="display:inline-block;min-width:860px;max-width:1100px;margin:16px;border-radius:10px;overflow:hidden;box-shadow:0 2px 12px #0006;font-family:'DejaVu Sans Mono',Menlo,monospace;font-size:14px">
<div style="background:#1e2a4a;padding:8px 14px;color:#9fb0d6">● ● ●&nbsp;&nbsp;terminal</div>
<pre style="margin:0;padding:14px;background:#111a33;color:#dbe4fb;white-space:pre;line-height:1.45"><span style="color:#fbbf24">$</span> <b>${esc(cmd.replace(/^\$ /, ''))}</b>\n${esc(resto.join('\n'))}</pre></div></body>`;
  const page = await browser.newPage({viewport: {width: 1160, height: 400}});
  await page.setContent(html);
  const el = await page.$('div');
  await el.screenshot({path: `${saida}/cli-${arq.replace(/\.txt$/, '')}.png`});
  await page.close();
}
await browser.close();
