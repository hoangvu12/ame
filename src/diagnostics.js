import { WS_URL } from './constants';

// A normal browser download: no upload and no WebSocket round trip needed.
export function downloadDiagnosticZIP() {
  const a = document.createElement('a');
  a.href = `${WS_URL.replace(/^ws:/, 'http:')}/diagnostics.zip`;
  a.download = 'ame-diagnostics.zip';
  document.body.appendChild(a);
  a.click();
  a.remove();
}
