import { exportDiagnostics } from './websocket';
import { toastPromise } from './toast';

// The embedded League browser can silently ignore anchor/blob downloads.
// Save through Ame instead and reveal the real file in Windows Explorer.
export function downloadDiagnosticZIP() {
  const pending = exportDiagnostics();
  toastPromise(pending, {
    loading: 'Saving logs…',
    success: (result) => `Logs saved: ${result.path}`,
    error: (error) => error.message || 'Unable to save logs. Try again.',
  });
  pending.catch(() => {});
  return pending;
}
