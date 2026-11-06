// Thin wrapper around the League client Toast API.
// Guards against Toast being unavailable (e.g. during tests or early init).

export function toastError(msg) {
  if (typeof Toast !== 'undefined') Toast.error(msg);
}

export function toastSuccess(msg) {
  if (typeof Toast !== 'undefined') Toast.success(msg);
}

export function toastPromise(promise, opts) {
  if (typeof Toast !== 'undefined') Toast.promise(promise, opts);
}
