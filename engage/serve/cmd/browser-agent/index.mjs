import http from 'node:http';
import { chromium } from 'playwright';

const listen = process.env.ENGAGE_BROWSER_LISTEN || ':8910';
const port = Number.parseInt(listen.split(':').pop() || '8910', 10);

async function execBrowser(target, wantScreenshot) {
  const browser = await chromium.launch({ headless: true });
  const page = await browser.newPage();
  let resp;
  try {
    resp = await page.goto(target, { waitUntil: 'domcontentloaded', timeout: 60_000 });
  } catch (err) {
    await browser.close();
    return { stdout: '', stderr: String(err), exit_code: 1, error: String(err) };
  }
  const title = await page.title();
  const status = resp?.status() ?? 0;
  const url = page.url();
  let screenshot = '';
  if (wantScreenshot) {
    screenshot = (await page.screenshot({ type: 'png' })).toString('base64');
  }
  await browser.close();
  const payload = { title, status, url, screenshot };
  return {
    stdout: `${JSON.stringify(payload)}\n`,
    stderr: '',
    exit_code: 0,
    error: '',
  };
}

function readBody(req) {
  return new Promise((resolve, reject) => {
    const chunks = [];
    req.on('data', (c) => chunks.push(c));
    req.on('end', () => resolve(Buffer.concat(chunks).toString('utf8')));
    req.on('error', reject);
  });
}

const server = http.createServer(async (req, res) => {
  try {
    if (req.url === '/health' && req.method === 'GET') {
      res.writeHead(200, { 'Content-Type': 'application/json' });
      res.end(JSON.stringify({ ok: true, service: 'engage-browser-agent', engine: 'playwright' }));
      return;
    }
    if (req.url === '/exec' && req.method === 'POST') {
      const raw = await readBody(req);
      const data = raw ? JSON.parse(raw) : {};
      const target = (data.target || (data.args && data.args[0]) || '').trim();
      if (!target) {
        res.writeHead(400, { 'Content-Type': 'text/plain' });
        res.end('target required');
        return;
      }
      const out = await execBrowser(target, Boolean(data.screenshot));
      res.writeHead(200, { 'Content-Type': 'application/json' });
      res.end(JSON.stringify(out));
      return;
    }
    res.writeHead(404);
    res.end('not found');
  } catch (err) {
    res.writeHead(500, { 'Content-Type': 'application/json' });
    res.end(JSON.stringify({ stdout: '', stderr: String(err), exit_code: 1, error: String(err) }));
  }
});

server.listen(port, '0.0.0.0', () => {
  console.log(`engage browser-agent (playwright) on :${port}`);
});
